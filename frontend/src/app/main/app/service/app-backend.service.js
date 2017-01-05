(function () {
  'use strict';
  angular.module('swan')
    .factory('appBackend', appBackend);

  /** @ngInject */
  function appBackend($resource) {
    return {
      app: app
    };

    function app(data) {
      var paramObj = data || {};
      return $resource(BACKEND_URL_BASE.defaultBase + '/v1/apps/:appId', {
        appId: paramObj.appId
      });
    }
  }
})();
