(function () {
  'use strict';

  angular
    .module('swan')
    .controller('UserController', UserController);

  /** @ngInject */
  function UserController(userBackend, $stateParams) {
    var params = {
      user: $stateParams.user
    };

    var vm = this;
    vm.apps = [];

    activate();

    function activate() {
      listApp()
    }

    function listApp() {
      userBackend.apps(params).query(function (data) {
        vm.apps = data;
      })
    }
  }
})();
