(function () {
  'use strict';

  angular
    .module('swan')
    .controller('ClusterController', ClusterController);

  /** @ngInject */
  function ClusterController(clusterBackend) {
    var vm = this;
    vm.cluster = {};

    activate();

    function activate() {
      getClusterInfo()
    }

    function getClusterInfo() {
      clusterBackend.cluster().get(function (data) {
        vm.cluster = data;
      })
    }
  }
})();
