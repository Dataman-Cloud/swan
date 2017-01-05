(function () {
  'use strict';
  angular.module('swan')
    .factory('userBackend', userBackend);

  /** @ngInject */
  function userBackend($resource) {
    return {
      apps: apps
    };

    function apps(data) {
      var paramObj = data || {};
      return $resource(BACKEND_URL_BASE.defaultBase + '/v_beta/apps', {
        fields: 'runAs=' + paramObj.user
      });
    }
  }
})();
