(function () {
  'use strict';

  angular
    .module('swan')
    .config(routerConfig);

  /** @ngInject */
  function routerConfig($stateProvider, $urlRouterProvider) {
    $stateProvider
      .state('home', {
        templateUrl: 'app/main/main.html',
        controller: 'MainController',
        controllerAs: 'main'
      })
      .state('home.cluster', {
        url: '/cluster',
        templateUrl: 'app/main/cluster/cluster.html',
        controller: 'ClusterController',
        controllerAs: 'cluster'
      })
      .state('home.user', {
        url: '/user',
        templateUrl: 'app/main/user/user.html',
        controller: 'UserController',
        controllerAs: 'user'
      })
      .state('home.app', {
        url: '/app',
        templateUrl: 'app/main/app/app.html',
        controller: 'AppController',
        controllerAs: 'app'
      })
      .state('home.task', {
        url: '/task',
        templateUrl: 'app/main/task/task.html',
        controller: 'TaskController',
        controllerAs: 'task'
      });

    $urlRouterProvider.otherwise(function ($injector) {
      var $state = $injector.get('$state');
      $state.go('home.cluster');
    });
  }

})();
