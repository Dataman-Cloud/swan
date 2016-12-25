(function () {
  'use strict';

  describe('controllers', function () {
    var vm;
    var clusterBackend;

    beforeEach(module('swan'));
    beforeEach(inject(function (_$controller_, _clusterBackend_) {
      spyOn(_clusterBackend_, 'cluster').and.callFake(function () {
        return {
          get: function () {
            return {
              "appCount": 3,
              "taskCount": 6,
              "cpuTotalOffered": 0.060000000000000005,
              "memTotalOffered": 30,
              "appStats": {
                "group1": 2,
                "xychu": 1
              }
            }
          }
        };
      });

      vm = _$controller_('ClusterController');
      clusterBackend = _clusterBackend_;
    }));

    it('should define vm.cluster object', function () {
      expect(angular.isObject(vm.cluster)).toBeTruthy();
      expect(vm.cluster.appStats).toBeTruthy();
    });
  });
})();
