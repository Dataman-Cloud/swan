(function () {
  'use strict';

  describe('service clusterBackend', function () {
    var clusterBackend;
    var $httpBackend;

    beforeEach(module('swan'));
    beforeEach(inject(function (_clusterBackend_, _$httpBackend_, $templateCache) {
      $templateCache.put('app/main/main.html', '.<template-goes-here />');
      $templateCache.put('app/main/cluster/cluster.html', '.<template-goes-here />');
      clusterBackend = _clusterBackend_;
      $httpBackend = _$httpBackend_;
    }));

    it('should be registered', function () {
      expect(clusterBackend).not.toEqual(null);
    });

    describe('clusterBackend function', function () {
      it('should exist', function () {
        expect(clusterBackend.cluster).not.toEqual(null);
      });

      it('should return data', function () {
        $httpBackend.when('GET', BACKEND_URL_BASE.defaultBase + '/stats').respond(200, {
          "appCount": 3,
          "taskCount": 6,
          "cpuTotalOffered": 0.060000000000000005,
          "memTotalOffered": 30,
          "appStats": {
            "group1": 2,
            "xychu": 1
          }
        });
        var data = clusterBackend.cluster().get();

        $httpBackend.flush();


        expect(data.appStats).toEqual(jasmine.any(Object));
      });


      it('should return a error', function () {
        $httpBackend.when('GET', BACKEND_URL_BASE.defaultBase + '/stats').respond(500, 'XHR Failed for');
        clusterBackend.cluster().get();
        $httpBackend.flush();

      });
    });
  });
})();
