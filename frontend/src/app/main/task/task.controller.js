(function () {
  'use strict';

  angular
    .module('swan')
    .controller('TaskController', TaskController);

  /** @ngInject */
  function TaskController(taskBackend, $stateParams, moment) {
    var params = {
      appId: $stateParams.app,
      taskIndex: $stateParams.task
    };

    var vm = this;
    vm.task = {};
    vm.to = moment().unix() * 1000;
    vm.from = moment().subtract(120, 'minutes').unix() * 1000;

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
