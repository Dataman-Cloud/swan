(function () {
  'use strict';
  angular.module('swan')
    .factory('clusterBackend', clusterBackend);

  /** @ngInject */
  function clusterBackend($resource) {
    return {
      cluster: cluster
    };

    function cluster() {
      return $resource(BACKEND_URL_BASE.defaultBase + '/stats');
    }
  }
})();
