// Nord Theme for ECharts
// Register both dark and light variants

(function() {
    var nordColors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];

    echarts.registerTheme('nord-dark', {
        color: nordColors,
        backgroundColor: 'transparent',
        textStyle: {color: '#D8DEE9'},
        title: {textStyle: {color: '#ECEFF4'}},
        legend: {textStyle: {color: '#D8DEE9'}},
        tooltip: {
            backgroundColor: '#3B4252',
            borderColor: '#4C566A',
            textStyle: {color: '#D8DEE9'},
        },
        categoryAxis: {
            axisLine: {lineStyle: {color: '#4C566A'}},
            axisTick: {lineStyle: {color: '#4C566A'}},
            axisLabel: {color: '#D8DEE9'},
            splitLine: {lineStyle: {color: '#3B4252'}},
        },
        valueAxis: {
            axisLine: {lineStyle: {color: '#4C566A'}},
            axisTick: {lineStyle: {color: '#4C566A'}},
            axisLabel: {color: '#D8DEE9'},
            splitLine: {lineStyle: {color: '#3B4252'}},
        },
        label: {color: '#D8DEE9'},
    });

    echarts.registerTheme('nord-light', {
        color: nordColors,
        backgroundColor: 'transparent',
        textStyle: {color: '#2E3440'},
        title: {textStyle: {color: '#2E3440'}},
        legend: {textStyle: {color: '#2E3440'}},
        tooltip: {
            backgroundColor: '#ECEFF4',
            borderColor: '#D8DEE9',
            textStyle: {color: '#2E3440'},
        },
        categoryAxis: {
            axisLine: {lineStyle: {color: '#D8DEE9'}},
            axisTick: {lineStyle: {color: '#D8DEE9'}},
            axisLabel: {color: '#2E3440'},
            splitLine: {lineStyle: {color: '#E5E9F0'}},
        },
        valueAxis: {
            axisLine: {lineStyle: {color: '#D8DEE9'}},
            axisTick: {lineStyle: {color: '#D8DEE9'}},
            axisLabel: {color: '#2E3440'},
            splitLine: {lineStyle: {color: '#E5E9F0'}},
        },
        label: {color: '#2E3440'},
    });
})();
