(function () {
  'use strict';
  angular.module('swan')
    .factory('httpInterceptor', httpInterceptor);

  /** @ngInject */
  function httpInterceptor($q, $injector) {
    //injector ui-notification
    var notification = null;
    var getNotification = function () {
      if (!notification) {
        notification = $injector.get('Notification');
      }
      return notification;
    };

    return {
      // optional method
      'request': function (config) {
        // do something on success
        return config;
      },
      // optional method
      'requestError': function (rejection) {
        // do something on error
        return $q.reject(rejection);
      },

      // optional method
      'response': function (response) {
        // do something on success
        return response;
      },

      // optional method
      'responseError': function (rejection) {
        // do something on error
        var msg = rejection.data ? rejection.data : '连接后端服务器异常。' + '</br>' + '请确认配置: ' + BACKEND_URL_BASE.defaultBase;
        getNotification().error(msg);
        return $q.reject(rejection);
      }

    }
  }
})();
