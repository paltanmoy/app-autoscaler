package sqldb

import (
	"database/sql"
	"time"

	"code.cloudfoundry.org/lager"
	_ "github.com/lib/pq"

	"autoscaler/db"
	"autoscaler/models"
)

type LockSQLDB struct {
	url    string
	logger lager.Logger
	sqldb  *sql.DB
}

func NewLockSQLDB(url string, logger lager.Logger) (*LockSQLDB, error) {
	sqldb, err := sql.Open(db.PostgresDriverName, url)
	if err != nil {
		logger.Error("open-lock-db", err, lager.Data{"url": url})
		return nil, err
	}

	err = sqldb.Ping()
	if err != nil {
		sqldb.Close()
		logger.Error("ping-lock-db", err, lager.Data{"url": url})
		return nil, err
	}

	return &LockSQLDB{
		url:    url,
		logger: logger,
		sqldb:  sqldb,
	}, nil
}

func (ldb *LockSQLDB) Close() error {
	err := ldb.sqldb.Close()
	if err != nil {
		ldb.logger.Error("Close-lock-db", err, lager.Data{"url": ldb.url})
		return err
	}
	return nil
}

func (ldb *LockSQLDB) FetchLock(locktype string) (lock models.Lock, err error) {
	ldb.logger.Info("Fetching locks ", lager.Data{"type": locktype})
	var (
		res_type      string
		res_owner     string
		res_timestamp int64
		res_ttl       time.Duration
	)
	query := "SELECT * FROM locks WHERE type = $1"
	fetchLockErr := ldb.sqldb.QueryRow(query, locktype).Scan(&res_owner, &res_type, &res_timestamp, &res_ttl)
	switch {
	case fetchLockErr == sql.ErrNoRows:
		ldb.logger.Info("No lock entry found", lager.Data{"type": locktype})
		return models.Lock{}, fetchLockErr
	case fetchLockErr != nil:
		ldb.logger.Error("Error occurs during lock fetching", fetchLockErr)
		return models.Lock{}, fetchLockErr
	default:
		ldb.logger.Info("Lock already exist", lager.Data{"Owner": res_owner, "Type": res_type, "Last_Modified_Timestamp": res_timestamp, "Ttl": res_ttl})
		lock := models.Lock{Owner: res_owner, Type: res_type, Last_Modified_Timestamp: res_timestamp, Ttl: res_ttl}
		return lock, nil
	}
}

func (ldb *LockSQLDB) ClaimLock(lockDetails models.Lock) (claimed bool, err error) {
	ldb.logger.Info("No lock owner found! claiming the lock", lager.Data{"Owner": lockDetails.Owner, "Type": lockDetails.Type, "Last_Modified_Timestamp": lockDetails.Last_Modified_Timestamp, "Ttl": lockDetails.Ttl})
	isClaimed := true
	tx, err := ldb.sqldb.Begin()
	if err != nil {
		ldb.logger.Error("Error starting  transaction", err)
		return false, err
	}
	defer func() {
		if err != nil {
			ldb.logger.Error("Rolling back claim lock transaction!", err)
			err = tx.Rollback()
			if err != nil {
				ldb.logger.Error("Failed to Rollback", err)
			}
			isClaimed = false
			return
		}
		err = tx.Commit()
		if err != nil {
			ldb.logger.Error("Failed to commit claimlock transaction", err)
			isClaimed = false
		}
	}()
	if _, err = tx.Exec("SELECT * FROM locks FOR UPDATE"); err != nil {
		ldb.logger.Error("Error during select for update", err)
		isClaimed = false
		return isClaimed, err
	}
	query := "INSERT INTO locks (owner,type,lock_timestamp,ttl) VALUES ($1,$2,$3,$4)"
	if _, err = tx.Exec(query, lockDetails.Owner, lockDetails.Type, lockDetails.Last_Modified_Timestamp, int64(lockDetails.Ttl/time.Second)); err != nil {
		ldb.logger.Error("Failed to insert lock details", err)
		isClaimed = false
		return isClaimed, err
	}
	return isClaimed, err
}

func (ldb *LockSQLDB) RenewLock(owner string) error {
	ldb.logger.Info("Renewing lock ", lager.Data{"Owner": owner})
	tx, err := ldb.sqldb.Begin()
	if err != nil {
		ldb.logger.Error("Error starting  transaction", err)
		return err
	}
	defer func() {
		if err != nil {
			ldb.logger.Error("Rolling back renew lock transaction!", err)
			err = tx.Rollback()
			if err != nil {
				ldb.logger.Error("Failed to Rollback", err)
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			ldb.logger.Error("Failed to commit renew transaction", err)
		}
	}()
	query := "SELECT * FROM locks where owner=$1 FOR UPDATE"
	if _, err = tx.Exec(query, owner); err != nil {
		ldb.logger.Error("Error during select for update", err)
		return err
	}
	updatequery := "UPDATE locks SET lock_timestamp=$1 where owner=$2"
	if _, err = tx.Exec(updatequery, time.Now().Unix(), owner); err != nil {
		ldb.logger.Error("Failed to update lock details", err)
		return err
	}
	return err
}

func (ldb *LockSQLDB) ReleaseLock(owner string) error {
	ldb.logger.Info("Releasing the lock", lager.Data{"Owner": owner})
	tx, err := ldb.sqldb.Begin()
	if err != nil {
		ldb.logger.Error("Error starting  transaction", err)
		return err
	}
	defer func() {
		if err != nil {
			ldb.logger.Error("Rolling back release lock transaction!", err)
			err = tx.Rollback()
			if err != nil {
				ldb.logger.Error("Failed to Rollback", err)
			}
			return
		}
		err = tx.Commit()
		if err != nil {
			ldb.logger.Error("Error during commit", err)
		}
	}()
	if _, err := tx.Exec("SELECT * FROM locks FOR UPDATE"); err != nil {
		ldb.logger.Error("Error during select for update", err)
		return err
	}
	query := "DELETE FROM locks WHERE owner = $1"
	if _, err := tx.Exec(query, owner); err != nil {
		ldb.logger.Error("Failed to delete lock details", err)
		return err
	}
	return nil
}

func (ldb *LockSQLDB) AcquireLock(lock models.Lock) (bool, error) {
	fetchedLock, err := ldb.FetchLock(lock.Type)
	if err != nil && err == sql.ErrNoRows {
		ldb.logger.Info("No one holds the lock, lets claim!")
		_, err = ldb.ClaimLock(lock)
		if err != nil {
			ldb.logger.Error("Failed to claim the lock", err)
			return false, err
		}
	} else if err != nil && err != sql.ErrNoRows {
		ldb.logger.Error("Failed to fetch lock", err)
		return false, err
	} else {
		if fetchedLock.Owner == lock.Owner && fetchedLock.Type == lock.Type {
			err = ldb.RenewLock(lock.Owner)
			if err != nil {
				ldb.logger.Error("Failed to renew lock", err)
				return false, err
			}
		} else {
			ldb.logger.Info("Someone else is the Owner", lager.Data{"Owner": fetchedLock.Owner, "Type": fetchedLock.Type})
			lastUpdatedTime := time.Unix(fetchedLock.Last_Modified_Timestamp, 0)
			if lastUpdatedTime.Add(time.Second * time.Duration(fetchedLock.Ttl)).Before(time.Now()) {
				ldb.logger.Info("Lock not renewed! Lets forcefully grab the lock")
				err = ldb.ReleaseLock(fetchedLock.Owner)
				if err != nil {
					ldb.logger.Error("Failed to release lock forcefully", err)
					return false, err
				}
				_, err = ldb.ClaimLock(lock)
				if err != nil {
					ldb.logger.Error("Failed to claim lock", err)
					return false, err
				}
			} else {
				ldb.logger.Info("Lock renewed and hold by owner")
				return false, nil
			}
		}
	}
	return true, nil
}
