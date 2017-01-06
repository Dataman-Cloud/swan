(function () {
  'use strict';

  angular
    .module('swan')
    .controller('AppController', AppController);

  /** @ngInject */
  function AppController(appBackend, $stateParams, moment) {
    var params = {appId: $stateParams.app};

    var vm = this;
    vm.end = moment().unix();
    vm.start = moment().subtract(120, 'minutes').unix();
    vm.app = {};

    activate();

    function activate() {
      getAppInfo();
      listAppEvents()
    }

    function getAppInfo() {
      if (params.appId) {
        appBackend.app(params).get(function (data) {
          vm.app = data;
        });
      }
    }

    function listAppEvents() {
      var source = new EventSource(BACKEND_URL_BASE.defaultBase + '/events');
      source.onmessage = function (e) {
        console.log(e.data)
      }
    }
  }
})();
