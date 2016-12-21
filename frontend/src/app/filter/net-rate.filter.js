(function () {
    'use strict';
    angular.module('swan')
        .filter('netRate', netRate);

    /* @ngInject */
    function netRate() {
        //////
        return function (rawSize) {
            rawSize = parseFloat(rawSize);
            var units = ['B/s', 'KB/s', 'MB/s'];
            var unitIndex = 0;
            while (rawSize >= 1024 && unitIndex < units.length - 1) {
                rawSize /= 1024;
                unitIndex++;
            }
            return rawSize.toFixed(2) + units[unitIndex];
        }
    }
})();
