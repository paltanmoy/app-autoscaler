package sqldb_test

import (
	. "autoscaler/db/sqldb"
	"autoscaler/models"
	"os"

	"code.cloudfoundry.org/lager"

	"time"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LockSqldb", func() {
	var (
		ldb             *LockSQLDB
		url             string
		logger          lager.Logger
		err             error
		fetchedLocks    models.Lock
		isClaimed       bool
		isLockAccquired bool
		testTTL         time.Duration
	)

	BeforeEach(func() {
		logger = lager.NewLogger("lock-sqldb-test")
		url = os.Getenv("DBURL")
		testTTL = time.Duration(30) * time.Second
	})

	Describe("NewLockSQLDB", func() {
		JustBeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
		})

		AfterEach(func() {
			if ldb != nil {
				err = ldb.Close()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		Context("when lock db url is not correct", func() {
			BeforeEach(func() {
				url = "postgres://not-exist-user:not-exist-password@localhost/autoscaler?sslmode=disable"
			})
			It("should error", func() {
				Expect(err).To(BeAssignableToTypeOf(&pq.Error{}))
			})

		})

		Context("when lock db url is correct", func() {
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ldb).NotTo(BeNil())
			})
		})
	})

	Describe("FetchLocks", func() {
		BeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanLockTable()
		})

		AfterEach(func() {
			err = ldb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			fetchedLocks, err = ldb.FetchLock("metrics_collector")
		})

		Context("when lock table is empty", func() {
			It("returns no lock details", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when policy table is not empty", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: 30}
				insertLockDetails(newlock)
			})

			It("returns all app ids", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fetchedLocks.Owner).To(Equal("21dd89da-9fcd-4dd2-8440-8a2f93c76d22"))
				Expect(fetchedLocks.Type).To(Equal("metrics_collector"))
				Expect(fetchedLocks.Ttl).To(Equal(time.Duration(30)))
			})
		})
	})

	Describe("ClaimLock", func() {
		BeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanLockTable()
		})

		AfterEach(func() {
			err = ldb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no lock owner found", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: 30}
				isClaimed, err = ldb.ClaimLock(newlock)
			})
			It("lock claimed successfully", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isClaimed).To(BeTrue())
			})
		})
	})

	Describe("Accquirelock", func() {
		BeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanLockTable()
		})

		AfterEach(func() {
			err = ldb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When nobody owns the lock", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(newlock)
			})

			It("Successfully accquire the lock", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isLockAccquired).To(BeTrue())
			})
		})

		Context("When nobody owns the lock of same type but other type locks hold by someone else", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "other_type", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(newlock)
			})

			It("Successfully accquire the lock", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isLockAccquired).To(BeTrue())
			})
		})

		Context("When lock owned by self", func() {
			BeforeEach(func() {
				firstlock := models.Lock{Owner: "089a4f3f-86c0-4c50-a563-29c983f91d61", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(firstlock)
				Expect(err).NotTo(HaveOccurred())
				secondlock := models.Lock{Owner: "089a4f3f-86c0-4c50-a563-29c983f91d61", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(secondlock)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should successfully renew the lock", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isLockAccquired).To(BeTrue())
			})
		})
		Context("Lock owned and recently renewed by someone else", func() {
			BeforeEach(func() {
				firstlock := models.Lock{Owner: "089a4f3f-86c0-4c50-a563-29c983f91d61", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(firstlock)
				Expect(err).NotTo(HaveOccurred())
				secondlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(secondlock)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Failed to accquire the lock", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isLockAccquired).NotTo(BeTrue())
			})
		})

		Context("Lock owned by someone else but not renewed", func() {
			BeforeEach(func() {
				firstlock := models.Lock{Owner: "089a4f3f-86c0-4c50-a563-29c983f91d61", Type: "metrics_collector", Last_Modified_Timestamp: 1498544155, Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(firstlock)
				Expect(err).NotTo(HaveOccurred())
				secondlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isLockAccquired, err = ldb.AcquireLock(secondlock)
				Expect(isLockAccquired).To(BeTrue())
			})
			It("Should successfully accquire the lock", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(isLockAccquired).To(BeTrue())
			})
		})
	})

	Describe("Renew Locks", func() {
		BeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanLockTable()
		})

		AfterEach(func() {
			err = ldb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When lock exists", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isClaimed, err = ldb.ClaimLock(newlock)
				Expect(err).NotTo(HaveOccurred())
				Expect(isClaimed).To(BeTrue())
				err = ldb.RenewLock("21dd89da-9fcd-4dd2-8440-8a2f93c76d22")
			})
			It("Should able to renew the lock Successfully", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Release Locks", func() {
		BeforeEach(func() {
			ldb, err = NewLockSQLDB(url, logger)
			Expect(err).NotTo(HaveOccurred())
			cleanLockTable()
		})

		AfterEach(func() {
			err = ldb.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When lock exists", func() {
			BeforeEach(func() {
				newlock := models.Lock{Owner: "21dd89da-9fcd-4dd2-8440-8a2f93c76d22", Type: "metrics_collector", Last_Modified_Timestamp: time.Now().Unix(), Ttl: testTTL}
				isClaimed, err = ldb.ClaimLock(newlock)
				Expect(err).NotTo(HaveOccurred())
				Expect(isClaimed).To(BeTrue())
				err = ldb.ReleaseLock("21dd89da-9fcd-4dd2-8440-8a2f93c76d22")
			})
			It("Should able to release the lock Successfully", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

})
