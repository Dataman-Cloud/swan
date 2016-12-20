(function () {
  'use strict';

  angular
    .module('swan')
    .controller('TaskController', TaskController);

  /** @ngInject */
  function TaskController(taskBackend, $stateParams) {
    var params = {
      appId: $stateParams.appId,
      taskIndex: $stateParams.taskIndex
    };

    var vm = this;
    vm.task = {};

    activate();

    function activate() {
      if (params.appId && params.taskIndex) {
        getTaskInfo()
      }
    }

    function getTaskInfo() {
      taskBackend.task(params).get(function (data) {
        vm.task = data;
      })
    }
  }
})();
