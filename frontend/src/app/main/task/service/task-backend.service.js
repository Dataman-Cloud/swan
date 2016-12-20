(function () {
  'use strict';
  angular.module('swan')
    .factory('taskBackend', taskBackend);

  /** @ngInject */
  function taskBackend($resource) {
    return {
      task: task
    };

    function task(data) {
      var paramObj = data || {};
      return $resource(BACKEND_URL_BASE.defaultBase + '/v_beta/apps/:appId/tasks/:taskIndex', {
        appId: paramObj.appId,
        taskIndex: paramObj.taskIndex
      });
    }
  }
})();
