'use strict';
var path = require('path');
module.exports = function(settingsObj) {
  var db = function(dbUri) {
    if (dbUri != null) {
      let uri = dbUri.replace(/\/$/g, "");
      let name = uri.slice(uri.lastIndexOf("/") + 1, uri.length);
      let server = uri.slice(0, uri.lastIndexOf("/"));
      return {
        uri: uri,
        name: name,
        server: server
      };
    }
  };

  var cleanupURI = function(apiServerUri) {
    if (apiServerUri != null) {
      return apiServerUri.replace(/\/$/g, "").toLowerCase();
    }
  };
  settingsObj.apiserver.uri = cleanupURI(settingsObj.apiserver.uri);
  settingsObj.dashboardRedirectUri = cleanupURI(settingsObj.dashboardRedirectUri);
  settingsObj.customMetricsUrl = cleanupURI(settingsObj.customMetricsUrl);
  var settings = {
    publicPort: settingsObj.publicPort,
    port : settingsObj.port,
    username: settingsObj.username,
    password: settingsObj.password,
    apiserver: settingsObj.apiserver,
    httpRequestTimeout: settingsObj.httpRequestTimeout,
    tls: settingsObj.tls,
    publicTls: settingsObj.publicTls,
    serviceCatalogPath: settingsObj.serviceCatalogPath,
    dashboardRedirectUri: settingsObj.dashboardRedirectUri,
    customMetricsUrl: settingsObj.customMetricsUrl
  };
  if (settingsObj.db) {
    var dbObj = db(settingsObj.db.uri);
    settings.db = {
      maxConnections: settingsObj.db.maxConnections,
      minConnections: settingsObj.db.minConnections,
      idleTimeout: settingsObj.db.idleTimeout,
      uri: dbObj.uri,
      name: dbObj.name,
      server: dbObj.server
    }

  }
  var isMissing = function(value){
    return typeof(value) === "undefined" || value === null;
  }
  var isNumber = function(value){
    return typeof(value) === "number";
  }
  var isString = function(value){
    return typeof(value) === "string";
  }
  var isObject = function(value){
    return typeof(value) === "object";
  }
  settings.validate = function() {
    if (isMissing(settings.publicPort)) {
      return { valid: false, message: "publicPort is required" };
    }
    if (!isNumber(settings.publicPort)) {
      return { valid: false, message: "publicPort must be a number" };
    }
    if (settings.publicPort < 1 || settings.publicPort > 65535) {
      return { valid: false, message: "value of publicPort must between 1 and 65535" };
    }

    if (isMissing(settings.port)) {
      return { valid: false, message: "port is required" };
    }
    if (!isNumber(settings.port)) {
      return { valid: false, message: "port must be a number" };
    }
    if (settings.port < 1 || settings.port > 65535) {
      return { valid: false, message: "value of port must between 1 and 65535" };
    }

    if (isMissing(settings.username)) {
      return { valid: false, message: "username is required" };
    }
    if (!isString(settings.username)) {
      return { valid: false, message: "username must be a string" };
    }
    if (isMissing(settings.password)) {
      return { valid: false, message: "password is required" };
    }
    if (!isString(settings.password)) {
      return { valid: false, message: "password must be a string" };
    }
    if (isMissing(settings.db.maxConnections)) {
      return { valid: false, message: "db.maxConnections is required" };
    }
    if (!isNumber(settings.db.maxConnections)) {
      return { valid: false, message: "db.maxConnections must be a number" };
    }
    if (settings.db.maxConnections <= 0) {
      return { valid: false, message: "db.maxConnections must be greater than 0" };
    }
    if (isMissing(settings.db.minConnections)) {
      return { valid: false, message: "db.minConnections is required" };
    }
    if (!isNumber(settings.db.minConnections)) {
      return { valid: false, message: "db.minConnections must be a number" };
    }
    if (settings.db.minConnections < 0) {
      return { valid: false, message: "db.minConnections must be greater than or equal to 0" };
    }
    if (isMissing(settings.db.idleTimeout)) {
      return { valid: false, message: "db.idleTimeout is required" };
    }
    if (!isNumber(settings.db.idleTimeout)) {
      return { valid: false, message: "db.idleTimeout must be a number" };
    }
    if (settings.db.idleTimeout <= 0) {
      return { valid: false, message: "db.idleTimeout must be greater than 0" };
    }
    if (isMissing(settings.db.uri)) {
      return { valid: false, message: "db.uri is required" };
    }
    if (!isString(settings.db.uri)) {
      return { valid: false, message: "db.uri must be a string" };
    }
    if (isMissing(settings.httpRequestTimeout)) {
      return { valid: false, message: "httpRequestTimeout is required" };
    }
    if (!isNumber(settings.httpRequestTimeout)) {
      return { valid: false, message: "httpRequestTimeout must be a number" };
    }
    if (settings.httpRequestTimeout <= 0) {
      return { valid: false, message: "value of httpRequestTimeout must be greater than 0" };
    } 
    if (isMissing(settings.apiserver.uri)){
      return { valid: false, message: "apiserver.uri is required" };
    }
    if (!isString(settings.apiserver.uri)){
      return { valid: false, message: "apiserver.uri must be a string" };
    }
    if (!isMissing(settings.apiserver.tls)){
      if (!isObject(settings.apiserver.tls)){
        return { valid: false, message: "apiserver.tls must be an object" };
      }
      if (isMissing(settings.apiserver.tls.keyFile)) {
        return { valid: false, message: "apiserver.tls.keyFile is required" };
      }
      if (!isString(settings.apiserver.tls.keyFile)) {
        return { valid: false, message: "apiserver.tls.keyFile must be a string" };
      }
      if (isMissing(settings.apiserver.tls.certFile)) {
        return { valid: false, message: "apiserver.tls.certFile is required" };
      }
      if (!isString(settings.apiserver.tls.certFile)) {
        return { valid: false, message: "apiserver.tls.certFile must be a string" };
      }
      if (isMissing(settings.apiserver.tls.caCertFile)) {
        return { valid: false, message: "apiserver.tls.caCertFile is required" };
      }
      if (!isString(settings.apiserver.tls.caCertFile)) {
        return { valid: false, message: "apiserver.tls.caCertFile must be a string" };
      }
    } 
    if(!isMissing(settings.tls)){
      if(!isObject(settings.tls)){
        return { valid: false, message: "tls must be an object" };
      } 
      if (isMissing(settings.tls.keyFile)) {
        return { valid: false, message: "tls.keyFile is required" };
      }
      if (!isString(settings.tls.keyFile)) {
        return { valid: false, message: "tls.keyFile must be a string" };
      }
      if (isMissing(settings.tls.certFile)) {
        return { valid: false, message: "tls.certFile is required" };
      }
      if (!isString(settings.tls.certFile)) {
        return { valid: false, message: "tls.certFile must be a string" };
      }
      if (isMissing(settings.tls.caCertFile)) {
        return { valid: false, message: "tls.caCertFile is required" };
      }
      if (!isString(settings.tls.caCertFile)) {
        return { valid: false, message: "tls.caCertFile must be a string" };
      }
    }

    if(!isMissing(settings.publicTls)){
      if(!isObject(settings.publicTls)){
        return { valid: false, message: "publicTls must be an object" };
      } 
      if (isMissing(settings.publicTls.keyFile)) {
        return { valid: false, message: "publicTls.keyFile is required" };
      }
      if (!isString(settings.publicTls.keyFile)) {
        return { valid: false, message: "publicTls.keyFile must be a string" };
      }
      if (isMissing(settings.publicTls.certFile)) {
        return { valid: false, message: "publicTls.certFile is required" };
      }
      if (!isString(settings.publicTls.certFile)) {
        return { valid: false, message: "publicTls.certFile must be a string" };
      }
      if (isMissing(settings.publicTls.caCertFile)) {
        return { valid: false, message: "publicTls.caCertFile is required" };
      }
      if (!isString(settings.publicTls.caCertFile)) {
        return { valid: false, message: "publicTls.caCertFile must be a string" };
      }
    }

    if (isMissing(settings.serviceCatalogPath)) {
      return {valid: false, message: "serviceCatalogPath is required"}
    }
    if (!isString(settings.serviceCatalogPath)) {
      return {valid: false, message: "serviceCatalogPath must be a string"}
    }
    if (!isMissing(settings.dashboardRedirectUri) && !isString(settings.dashboardRedirectUri)) {
      return {valid: false, message: "dashboardRedirectUri must be a string"}
    }
    if (!isMissing(settings.customMetricsUrl) && !isString(settings.customMetricsUrl)) {
      return {valid: false, message: "customMetricsUrl must be a string"}
    }
    return { valid: true }
  }

  return settings;
};
