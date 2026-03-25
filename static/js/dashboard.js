document.addEventListener('alpine:init', () => {
    Alpine.data('dashboard', () => ({
        season: localStorage.getItem('db_season') || document.getElementById('dashboard-app').dataset.defaultSeason || '',
        games: [],
        positions: [],
        availableLeagues: [],
        availablePositions: [],
        filterLeague: localStorage.getItem('db_filterLeague') || '',
        filterPositions: JSON.parse(localStorage.getItem('db_filterPositions') || '[]'),
        loading: true,
        view: localStorage.getItem('db_view') || 'year',
        onlyCompleted: localStorage.getItem('db_onlyCompleted') === 'true',
        showRecentGames: localStorage.getItem('db_showRecentGames') === 'true',
        recentGamesLimit: localStorage.getItem('db_recentGamesLimit') || '10',
        today: new Date().toISOString().slice(0, 10),

        togglePosition(arr, p) {
            var i = arr.indexOf(p);
            if (i >= 0) arr.splice(i, 1); else arr.push(p);
        },

        get filtered() {
            return this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterLeague && g.league_id != this.filterLeague) return false;
                if (this.filterPositions.length
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
        },

        get stats() {
            var f = this.filtered;
            var fee = f.reduce((s, g) => s + g.fee, 0);
            var travel = f.reduce((s, g) => s + g.travel, 0);
            return {
                count: f.length,
                fee: fee,
                travel: travel,
                total: fee + travel,
                km: f.reduce((s, g) => s + g.km, 0),
            };
        },

        get upcomingGames() {
            return this.games
                .filter(g => g.date >= this.today)
                .sort((a, b) => a.date.localeCompare(b.date)
                    || (a.time || '').localeCompare(b.time || ''));
        },

        get recentGames() {
            var all = this.games
                .filter(g => {
                    if (g.date >= this.today) return false;
                    if (this.filterLeague && g.league_id != this.filterLeague) return false;
                    if (this.filterPositions.length
                        && !this.filterPositions.includes(g.position)) return false;
                    return true;
                })
                .sort((a, b) => b.date.localeCompare(a.date)
                    || (b.time || '').localeCompare(a.time || ''));
            var limit = parseInt(this.recentGamesLimit);
            return limit > 0 ? all.slice(0, limit) : all;
        },

        overviewGames: [],
        overviewPositions: [],
        overviewLeagues: [],
        overviewLoaded: false,
        ovFilterLeague: '',
        ovFilterPositions: [],
        ovYearFrom: '',
        ovYearTo: '',

        get allOverviewYears() {
            var seen = new Set();
            this.overviewGames.forEach(g => seen.add(g.year));
            return [...seen].sort().reverse();
        },

        get overviewFiltered() {
            return this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterLeague && g.league_id != this.ovFilterLeague) return false;
                if (this.ovFilterPositions.length
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
        },

        get overviewForPositions() {
            return this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterLeague && g.league_id != this.ovFilterLeague) return false;
                if (this.ovFilterPositions.length > 1
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
        },

        get overviewYears() {
            var seen = new Set();
            this.overviewFiltered.forEach(g => seen.add(g.year));
            return [...seen].sort();
        },

        overviewAggByYear(games) {
            var byYear = {};
            games.forEach(g => {
                if (!byYear[g.year]) byYear[g.year] = {
                    count: 0, fee: 0, travel: 0, km: 0,
                    by_position: {},
                };
                var y = byYear[g.year];
                y.count++;
                y.fee += g.fee;
                y.travel += g.travel;
                y.km += g.km;
                y.by_position[g.position] =
                    (y.by_position[g.position] || 0) + 1;
            });
            return byYear;
        },

        get overviewByYear() {
            return this.overviewAggByYear(this.overviewFiltered);
        },

        get overviewByYearForPositions() {
            return this.overviewAggByYear(this.overviewForPositions);
        },

        get overviewStats() {
            var f = this.overviewFiltered;
            var fee = f.reduce((s, g) => s + g.fee, 0);
            var travel = f.reduce((s, g) => s + g.travel, 0);
            return {
                count: f.length,
                fee: fee,
                travel: travel,
                total: fee + travel,
                km: f.reduce((s, g) => s + g.km, 0),
            };
        },

        _germanyMapLoaded: false,
        init() {
            this.setupEChartsResize();
            // Load Germany GeoJSON for ECharts
            if (!this._germanyMapLoaded && typeof echarts !== 'undefined') {
                fetch('/static/js/germany.geo.json')
                    .then(r => r.json())
                    .then(geo => {
                        echarts.registerMap('Germany', geo);
                        this._germanyMapLoaded = true;
                        // Re-render all maps now that geo is available
                        this.$nextTick(() => {
                            if (this.chartLib === 'echarts') {
                                if (this.games.length) {
                                    this.renderMap('chart-map', this.filtered);
                                    this.renderVenueTop('chart-venues-top', this.filtered);
                                }
                                if (this.view === 'overview' && this.overviewLoaded) {
                                    this.renderMap('chart-overview-map', this.overviewFiltered);
                                    this.renderVenueTop('chart-overview-venues-top', this.overviewFiltered);
                                }
                            }
                        });
                    });
            }
            this.loadSeason();
            if (this.view === 'overview') {
                this.loadOverview();
            }
            this.$watch('season', (v) => {
                localStorage.setItem('db_season', v);
                this.loadSeason();
            });
            this.$watch('filtered', () => this.renderCharts());
            this.$watch('filterLeague', (v) => {
                localStorage.setItem('db_filterLeague', v);
                this.renderPositionChart();
            });
            this.$watch('filterPositions', (v) => {
                localStorage.setItem('db_filterPositions', JSON.stringify(v));
            });
            this.$watch('onlyCompleted', (v) => {
                localStorage.setItem('db_onlyCompleted', v);
                this.renderCharts();
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('showRecentGames', (v) => {
                localStorage.setItem('db_showRecentGames', v);
            });
            this.$watch('recentGamesLimit', (v) => {
                localStorage.setItem('db_recentGamesLimit', v);
            });
            this.$watch('ovFilterLeague', (v) => {
                localStorage.setItem('db_ovFilterLeague', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovFilterPositions', (v) => {
                localStorage.setItem('db_ovFilterPositions', JSON.stringify(v));
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovYearFrom', (v) => {
                localStorage.setItem('db_ovYearFrom', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovYearTo', (v) => {
                localStorage.setItem('db_ovYearTo', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('view', (v) => {
                localStorage.setItem('db_view', v);
                if (v === 'overview') {
                    this.loadOverview();
                } else {
                    this.$nextTick(() => {
                        this.renderCharts();
                        setTimeout(() => this._resizeAllECharts(), 200);
                    });
                }
            });
            window.addEventListener('theme-changed', () => {
                this.disposeECharts();
                this.$nextTick(() => {
                    this.renderCharts();
                    if (this.view === 'overview') this.renderOverviewCharts();
                });
            });
        },

        _seasonLoaded: false,
        async loadSeason() {
            this.loading = true;
            if (this._seasonLoaded) {
                this.filterLeague = '';
                this.filterPositions = [];
            }
            this._seasonLoaded = true;
            var res = await fetch('/api/dashboard/season/' + this.season);
            var data = await res.json();
            this.games = data.games;
            this.positions = data.positions;
            this.availableLeagues = data.available_leagues;
            this.availablePositions = data.available_positions;
            this.loading = false;
            this.$nextTick(() => this.renderCharts());
        },

        renderCharts() {
            this.renderPositionChart();
            this.renderMonthlyChart();
            this.renderLeagueChart();
            this.renderCalendarHeatmap();
            this.renderFeeBar('chart-fee-total', g => g.fee + g.travel);
            this.renderFeeBar('chart-fee-base', g => g.fee);
            this.renderFeeBar('chart-fee-travel', g => g.travel);
            this.renderVenueTop('chart-venues-top', this.filtered);
            this.renderMap('chart-map', this.filtered);
        },

        // ECharts helper: init or reuse instance
        ecInit(chartId, height) {
            var el = document.getElementById(chartId);
            if (!el) return null;
            el.style.height = (height || 350) + 'px';
            var instance = echarts.getInstanceByDom(el);
            if (instance) instance.dispose();
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            return echarts.init(el, dark ? 'nord-dark' : 'nord-light');
        },

        // Resize all ECharts instances on window resize
        _resizeHandler: null,
        setupEChartsResize() {
            if (this._resizeHandler) return;
            this._resizeHandler = () => {
                document.querySelectorAll('[id^="chart-"]').forEach(el => {
                    var instance = echarts.getInstanceByDom(el);
                    if (instance) instance.resize();
                });
            };
            window.addEventListener('resize', this._resizeHandler);
        },

        _resizeAllECharts() {
            document.querySelectorAll('[id^="chart-"]').forEach(el => {
                var instance = echarts.getInstanceByDom(el);
                if (instance) instance.resize();
            });
        },

        // Dispose all ECharts instances (on lib switch or theme change)
        disposeECharts() {
            document.querySelectorAll('[id^="chart-"]').forEach(el => {
                var instance = echarts.getInstanceByDom(el);
                if (instance) instance.dispose();
            });
        },

        fmtDate(d) {
            if (!d || d.length < 10) return d;
            return d.slice(8,10) + '.' + d.slice(5,7) + '.' + d.slice(0,4);
        },

        eur(v) {
            return v.toFixed(2).replace('.', ',').replace(/\B(?=(\d{3})+(?!\d))/g, '.') + ' €';
        },


        // --- ECharts chart functions ---

        renderPositionChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterLeague && g.league_id != this.filterLeague) return false;
                if (this.filterPositions.length > 1
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
            var counts = {};
            games.forEach(g => { counts[g.position] = (counts[g.position] || 0) + 1; });
            var entries = Object.entries(counts).sort((a, b) => a[1] - b[1]);
            var labels = entries.map(e => e[0]);
            var values = entries.map(e => e[1]);

            var chart = this.ecInit('chart-positions');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                grid: {left: 40, right: 20, top: 10, bottom: 20, containLabel: true},
                xAxis: {type: 'value'},
                yAxis: {type: 'category', data: labels},
                series: [{
                    type: 'bar',
                    data: values.map((v, i) => ({value: v, itemStyle: {color: colors[i % colors.length]}})),
                    label: {show: true, position: 'right', fontSize: 11},
                }],
            });
        },

        renderMonthlyChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var allMonths = ['Jan','Feb','Mär','Apr','Mai','Jun','Jul','Aug','Sep','Okt','Nov','Dez'];
            var f = this.filtered;
            var raw = {};
            var activeMonths = new Set();
            f.forEach(g => {
                var m = parseInt(g.month) - 1;
                raw[g.position] = raw[g.position] || {};
                raw[g.position][m] = (raw[g.position][m] || 0) + 1;
                activeMonths.add(m);
            });
            var months = [...activeMonths].sort((a, b) => a - b);
            var monthLabels = months.map(m => allMonths[m]);
            var posWithData = this.positions.filter(p => raw[p]);

            // Calculate totals per month for stack labels
            var totals = months.map(m => {
                var t = 0;
                posWithData.forEach(p => { t += (raw[p] && raw[p][m]) || 0; });
                return t;
            });

            var chart = this.ecInit('chart-monthly');
            if (!chart) return;
            var series = posWithData.map((p, idx) => ({
                name: p,
                type: 'bar',
                stack: 'total',
                data: months.map(m => raw[p][m] || 0),
                itemStyle: {color: colors[idx % colors.length]},
                label: {show: false},
            }));
            // Show totals on top of last series
            if (series.length > 0) {
                series[series.length - 1].label = {
                    show: true,
                    position: 'top',
                    fontSize: 11,
                    formatter: function(params) { return totals[params.dataIndex]; },
                };
            }
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                legend: {data: posWithData, bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 30, right: 10, top: 20, bottom: 30, containLabel: true},
                xAxis: {type: 'category', data: monthLabels},
                yAxis: {type: 'value'},
                series: series,
            });
        },

        renderLeagueChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterPositions.length
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
            var leagueData = {};
            var posPresent = new Set();
            games.forEach(g => {
                leagueData[g.league] = leagueData[g.league] || {};
                leagueData[g.league][g.position] =
                    (leagueData[g.league][g.position] || 0) + 1;
                posPresent.add(g.position);
            });
            var leagues = Object.entries(leagueData)
                .map(([lg, pos]) => [lg, Object.values(pos).reduce((a, b) => a + b, 0)])
                .sort((a, b) => a[1] - b[1])
                .map(e => e[0]);
            var posOrder = this.positions.filter(p => posPresent.has(p));

            var leagueTotals = leagues.map(lg =>
                Object.values(leagueData[lg]).reduce((a, b) => a + b, 0)
            );

            var chart = this.ecInit('chart-leagues', Math.max(350, leagues.length * 25));
            if (!chart) return;
            var series = posOrder.map((p, i) => ({
                name: p,
                type: 'bar',
                stack: 'total',
                data: leagues.map(lg => (leagueData[lg] && leagueData[lg][p]) || 0),
                itemStyle: {color: colors[i % colors.length]},
            }));
            if (series.length > 0) {
                series[series.length - 1].label = {
                    show: true,
                    position: 'right',
                    fontSize: 10,
                    formatter: function(params) { return leagueTotals[params.dataIndex]; },
                };
            }
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                legend: {data: posOrder, orient: 'vertical', right: 0, top: 'middle', textStyle: {fontSize: 10}},
                grid: {left: 10, right: 80, top: 10, bottom: 10, containLabel: true},
                xAxis: {type: 'value'},
                yAxis: {type: 'category', data: leagues, axisLabel: {fontSize: 10}},
                series: series,
            });
        },

        renderTreemap(chartId, games) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';

            var posData = {};
            games.forEach(g => {
                if (!posData[g.position]) posData[g.position] = {};
                posData[g.position][g.league] = (posData[g.position][g.league] || 0) + 1;
            });

            var posOrder = this.positions.filter(p => posData[p]);
            var treeData = posOrder.map((pos, i) => {
                var leagues = Object.entries(posData[pos]).sort((a, b) => b[1] - a[1]);
                return {
                    name: pos + ' (' + leagues.reduce((s, e) => s + e[1], 0) + ')',
                    itemStyle: {color: colors[i % colors.length]},
                    children: leagues.map(([lg, count]) => ({
                        name: lg,
                        value: count,
                    })),
                };
            });

            var chart = this.ecInit(chartId);
            if (!chart) return;
            chart.setOption({
                tooltip: {
                    formatter: function(info) {
                        return info.name + ': ' + (info.value || '') + ' Spiele';
                    },
                },
                series: [{
                    type: 'treemap',
                    data: treeData,
                    top: 0, left: 0, right: 0, bottom: 25,
                    roam: false,
                    nodeClick: 'zoomToNode',
                    zoomToNodeRatio: 0.1 * 0.1,
                    drillDownIcon: '▶',
                    breadcrumb: {
                        show: true, bottom: 0, height: 20,
                        itemStyle: {color: dark ? '#3B4252' : '#D8DEE9'},
                        textStyle: {color: dark ? '#D8DEE9' : '#2E3440', fontSize: 11},
                    },
                    label: {show: true, formatter: '{b}\n{c}', fontSize: 11},
                    upperLabel: {
                        show: true, height: 18, fontSize: 11,
                        color: dark ? '#ECEFF4' : '#2E3440',
                        textBorderColor: 'transparent',
                        textBorderWidth: 0,
                        textShadowColor: 'transparent',
                        textShadowBlur: 0,
                    },
                    itemStyle: {
                        borderColor: dark ? '#2E3440' : '#ECEFF4',
                    },
                    levels: [
                        {
                            itemStyle: {borderWidth: 0, gapWidth: 5},
                            upperLabel: {show: true},
                        },
                        {
                            itemStyle: {gapWidth: 1},
                        },
                        {
                            colorSaturation: [0.35, 0.5],
                            itemStyle: {gapWidth: 1, borderColorSaturation: 0.6},
                        },
                    ],
                }],
            });
        },

        async loadOverview() {
            if (this.overviewLoaded) {
                this.$nextTick(() => this.renderOverviewCharts());
                return;
            }
            var res = await fetch('/api/dashboard/overview');
            var data = await res.json();
            this.overviewGames = data.games;
            this.overviewPositions = data.positions;
            this.overviewLeagues = data.available_leagues;
            this.ovFilterLeague = localStorage.getItem('db_ovFilterLeague') || '';
            this.ovFilterPositions = JSON.parse(
                localStorage.getItem('db_ovFilterPositions') || '[]'
            );
            var allYrs = this.allOverviewYears;
            var maxY = allYrs[0] || '';
            var minDefault = maxY ? '' + (parseInt(maxY) - 9) : '';
            this.ovYearFrom = localStorage.getItem('db_ovYearFrom') || minDefault;
            this.ovYearTo = localStorage.getItem('db_ovYearTo') || maxY;
            this.overviewLoaded = true;
            this.$nextTick(() => {
                this.renderOverviewCharts();
                setTimeout(() => this._resizeAllECharts(), 200);
            });
        },



        renderOverviewCharts() {
            if (!this.overviewLoaded) return;
            this.renderOverviewGamesPerYear();
            this.renderOverviewPositionTrend();
            this.renderOverviewPositionPie();
            this.renderOverviewLeagues();
            this.renderOverviewFeePerYear();
            this.renderOverviewAvgPerGame();
            this.renderOverviewKmPerYear();
            this.renderOverviewSankey();
            this.renderOverviewRiver();
            this.renderVenueTop('chart-overview-venues-top', this.overviewFiltered);
            this.renderMap('chart-overview-map', this.overviewFiltered);
        },

        // --- ECharts Overview Charts ---

        renderOverviewGamesPerYear() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var posWithData = this.overviewPositions.filter(p =>
                years.some(y => (byYear[y].by_position[p] || 0) > 0)
            );
            var totals = years.map(y => byYear[y].count);

            var chart = this.ecInit('chart-overview-games');
            if (!chart) return;
            var series = posWithData.map((p, i) => ({
                name: p,
                type: 'bar',
                stack: 'total',
                data: years.map(y => byYear[y].by_position[p] || 0),
                itemStyle: {color: colors[i % colors.length]},
            }));
            if (series.length > 0) {
                series[series.length - 1].label = {
                    show: true, position: 'top', fontSize: 10,
                    formatter: function(params) { return totals[params.dataIndex]; },
                };
            }
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                legend: {data: posWithData, bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 30, right: 10, top: 20, bottom: 30, containLabel: true},
                xAxis: {type: 'category', data: years},
                yAxis: {type: 'value'},
                series: series,
            });
        },

        renderOverviewPositionTrend() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var posWithData = this.overviewPositions.filter(p =>
                years.some(y => (byYear[y].by_position[p] || 0) > 0)
            );

            var chart = this.ecInit('chart-overview-trend');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', order: 'valueDesc'},
                legend: {data: posWithData, bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 30, right: 50, top: 10, bottom: 30, containLabel: true},
                xAxis: {
                    type: 'category', data: years,
                    boundaryGap: false,
                    splitLine: {show: true, lineStyle: {type: 'dashed'}},
                    axisLine: {show: false},
                    axisTick: {show: false},
                },
                yAxis: {
                    type: 'value',
                    splitLine: {show: false},
                    axisLine: {show: false},
                    axisTick: {show: false},
                    axisLabel: {show: false},
                },
                emphasis: {
                    focus: 'series',
                },
                series: posWithData.map((p, i) => ({
                    name: p,
                    type: 'line',
                    smooth: 0.6,
                    symbol: 'circle',
                    symbolSize: function(val) { return val === 0 ? 0 : 8; },
                    data: years.map(y => byYear[y].by_position[p] || 0),
                    lineStyle: {color: colors[i % colors.length], width: 3},
                    itemStyle: {color: colors[i % colors.length]},
                    emphasis: {
                        focus: 'series',
                        lineStyle: {width: 4},
                    },
                    blur: {
                        lineStyle: {opacity: 0.2, width: 2},
                        itemStyle: {opacity: 0.2},
                    },
                    endLabel: {
                        show: true,
                        formatter: '{a}',
                        fontSize: 10,
                        distance: 5,
                    },
                })),
            });
        },

        renderOverviewPositionPie() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYearForPositions;
            var totals = {};
            Object.values(byYear).forEach(y => {
                Object.entries(y.by_position).forEach(([p, c]) => {
                    totals[p] = (totals[p] || 0) + c;
                });
            });
            var sorted = Object.entries(totals).sort((a, b) => b[1] - a[1]);

            var chart = this.ecInit('chart-overview-pie');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'item', formatter: '{b}: {c} ({d}%)'},
                series: [{
                    type: 'pie',
                    radius: ['35%', '65%'],
                    center: ['50%', '50%'],
                    data: sorted.map((e, i) => ({
                        name: e[0], value: e[1],
                        itemStyle: {color: colors[i % colors.length]},
                    })),
                    label: {formatter: '{b}\n{d}%', fontSize: 11},
                    emphasis: {
                        label: {show: true, fontSize: 13, fontWeight: 'bold'},
                    },
                }],
            });
        },

        renderOverviewLeagues() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterPositions.length
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
            if (!games.length) return;

            var seenYears = new Set();
            games.forEach(g => seenYears.add(g.year));
            var years = [...seenYears].sort();

            var leagueTotals = {};
            games.forEach(g => { leagueTotals[g.league] = (leagueTotals[g.league] || 0) + 1; });
            var leagues = Object.keys(leagueTotals).sort((a, b) => leagueTotals[b] - leagueTotals[a]);

            var counts = {};
            games.forEach(g => {
                var key = g.league + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });

            var chart = this.ecInit('chart-overview-leagues');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                legend: {data: years, bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 10, right: 10, top: 10, bottom: 30, containLabel: true},
                xAxis: {type: 'category', data: leagues, axisLabel: {rotate: -45, fontSize: 9}},
                yAxis: {type: 'value'},
                series: years.map((y, i) => ({
                    name: y,
                    type: 'bar',
                    stack: 'total',
                    data: leagues.map(lg => counts[lg + '|' + y] || 0),
                    itemStyle: {color: colors[i % colors.length]},
                })),
            });
        },

        renderOverviewFeePerYear() {
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var fees = years.map(y => byYear[y] ? byYear[y].fee : 0);
            var travels = years.map(y => byYear[y] ? byYear[y].travel : 0);
            var totals = years.map((y, i) => fees[i] + travels[i]);

            var chart = this.ecInit('chart-overview-fee');
            if (!chart) return;
            var series = [
                {name: 'Pauschale', type: 'bar', stack: 'total', data: fees, itemStyle: {color: '#5E81AC'}},
                {name: 'Fahrtkosten', type: 'bar', stack: 'total', data: travels, itemStyle: {color: '#88C0D0'}},
            ];
            series[series.length - 1].label = {
                show: true, position: 'top', fontSize: 10,
                formatter: function(params) { return Math.round(totals[params.dataIndex]) + ' €'; },
            };
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}, valueFormatter: v => Math.round(v) + ' €'},
                legend: {data: ['Pauschale', 'Fahrtkosten'], bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 50, right: 10, top: 30, bottom: 30, containLabel: true},
                xAxis: {type: 'category', data: years},
                yAxis: {type: 'value', axisLabel: {formatter: v => Math.round(v) + ' €'}},
                series: series,
            });
        },

        renderOverviewAvgPerGame() {
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var avgs = years.map(y => {
                var d = byYear[y];
                if (!d || d.count === 0) return 0;
                return (d.fee + d.travel) / d.count;
            });

            var chart = this.ecInit('chart-overview-avg');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', valueFormatter: v => v.toFixed(0) + ' €'},
                grid: {left: 50, right: 10, top: 20, bottom: 10, containLabel: true},
                xAxis: {type: 'category', data: years},
                yAxis: {type: 'value', axisLabel: {formatter: v => v.toFixed(0) + ' €'}},
                series: [{
                    type: 'bar',
                    data: avgs,
                    itemStyle: {color: '#A3BE8C'},
                    label: {show: true, position: 'top', fontSize: 10, formatter: p => p.value === 0 ? '' : p.value.toFixed(0) + ' €'},
                }],
            });
        },

        renderOverviewKmPerYear() {
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var kms = years.map(y => byYear[y] ? byYear[y].km : 0);

            var chart = this.ecInit('chart-overview-km');
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', valueFormatter: v => v.toLocaleString('de-DE') + ' km'},
                grid: {left: 50, right: 10, top: 20, bottom: 10, containLabel: true},
                xAxis: {type: 'category', data: years},
                yAxis: {type: 'value', axisLabel: {formatter: v => v.toLocaleString('de-DE')}},
                series: [{
                    type: 'bar',
                    data: kms,
                    itemStyle: {color: '#81A1C1'},
                    label: {show: true, position: 'top', fontSize: 10, formatter: p => p.value === 0 ? '' : p.value.toLocaleString('de-DE')},
                }],
            });
        },

        renderOverviewRiver() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.overviewFiltered;
            if (!games.length) return;

            var years = this.overviewYears;
            var positions = this.overviewPositions.filter(p =>
                games.some(g => g.position === p)
            );

            // ThemeRiver data: [date, count, position]
            var data = [];
            var counts = {};
            games.forEach(g => {
                var key = g.year + '|' + g.position;
                counts[key] = (counts[key] || 0) + 1;
            });
            years.forEach(y => {
                positions.forEach(p => {
                    data.push([y + '-07-01', counts[y + '|' + p] || 0, p]);
                });
            });

            // Build totals per year for tooltip
            var yearTotals = {};
            years.forEach(y => {
                var total = 0;
                positions.forEach(p => { total += counts[y + '|' + p] || 0; });
                yearTotals[y] = total;
            });

            var chart = this.ecInit('chart-overview-river', 300);
            if (!chart) return;
            chart.setOption({
                tooltip: {
                    trigger: 'axis',
                    axisPointer: {type: 'line'},
                    formatter: function(params) {
                        if (!params || !params.length) return '';
                        var date = params[0].data[0];
                        var year = date.substring(0, 4);
                        var lines = ['<b>' + year + '</b> (Gesamt: ' + (yearTotals[year] || 0) + ')'];
                        params.sort((a, b) => b.data[1] - a.data[1]);
                        params.forEach(p => {
                            if (p.data[1] > 0) {
                                lines.push(p.marker + ' ' + p.data[2] + ': <b>' + p.data[1] + '</b>');
                            }
                        });
                        return lines.join('<br>');
                    },
                },
                legend: {
                    data: positions,
                    bottom: 0,
                    textStyle: {fontSize: 10},
                },
                singleAxis: {
                    type: 'time',
                    bottom: 40,
                    top: 20,
                    axisLabel: {fontSize: 10, formatter: '{yyyy}'},
                },
                series: [{
                    type: 'themeRiver',
                    data: data,
                    label: {
                        show: true, fontSize: 10,
                        formatter: function(params) {
                            return params.data[1] > 0 ? params.data[2] : '';
                        },
                    },
                    emphasis: {focus: 'self'},
                    itemStyle: {
                        color: function(params) {
                            var idx = positions.indexOf(params.data[2]);
                            return colors[idx >= 0 ? idx % colors.length : 0];
                        },
                    },
                }],
            });
        },

        renderOverviewSankey() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var games = this.overviewFiltered;
            if (!games.length) return;

            var years = this.overviewYears;
            // Use position sorter order (from API), filtered to those with data
            var positions = this.overviewPositions.filter(p =>
                games.some(g => g.position === p)
            ).reverse(); // reverse so first position is at top in heatmap

            var counts = {};
            games.forEach(g => {
                var key = g.position + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });

            // Heatmap data: [yearIdx, posIdx, count]
            var heatData = [];
            var max = 0;
            positions.forEach((p, pi) => {
                years.forEach((y, yi) => {
                    var val = counts[p + '|' + y] || 0;
                    heatData.push([yi, pi, val]);
                    if (val > max) max = val;
                });
            });

            var chart = this.ecInit('chart-overview-sankey', Math.max(300, positions.length * 45));
            if (!chart) return;
            chart.setOption({
                tooltip: {
                    formatter: function(params) {
                        return positions[params.value[1]] + ' ' + years[params.value[0]] + ': ' + params.value[2] + ' Spiele';
                    },
                },
                grid: {left: 50, right: 10, top: 10, bottom: 30, containLabel: true},
                xAxis: {type: 'category', data: years, splitArea: {show: true}},
                yAxis: {type: 'category', data: positions, splitArea: {show: true}},
                visualMap: {
                    min: 0, max: max || 1,
                    calculable: true,
                    orient: 'horizontal',
                    left: 'center', bottom: 0,
                    inRange: {
                        color: ['#3B4252', '#5E81AC', '#88C0D0', '#D08770', '#BF616A'],
                    },
                    textStyle: {color: dark ? '#D8DEE9' : '#2E3440', fontSize: 10},
                    show: false,
                },
                series: [{
                    type: 'heatmap',
                    data: heatData,
                    label: {
                        show: true,
                        formatter: function(params) { return params.value[2] || ''; },
                        fontSize: 11,
                        color: '#ECEFF4',
                        textBorderColor: '#2E3440',
                        textBorderWidth: 2,
                    },
                    itemStyle: {
                        borderColor: dark ? '#2E3440' : '#ECEFF4',
                        borderWidth: 2,
                        borderRadius: 3,
                    },
                }],
            });
        },

        renderCalendarHeatmap() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var posColorMap = {};
            this.positions.forEach((p, i) => { posColorMap[p] = colors[i % colors.length]; });

            var games = this.filtered;
            if (!games.length) return;

            // Group games by date
            var byDate = {};
            games.forEach(g => {
                if (!byDate[g.date]) byDate[g.date] = [];
                byDate[g.date].push(g);
            });

            // Calendar data: [date, count, dominantPosition]
            var calData = Object.entries(byDate).map(([date, gs]) => {
                // Dominant position = most frequent
                var posCounts = {};
                gs.forEach(g => { posCounts[g.position] = (posCounts[g.position] || 0) + 1; });
                var dominant = Object.entries(posCounts).sort((a, b) => b[1] - a[1])[0][0];
                return {
                    date: date,
                    count: gs.length,
                    position: dominant,
                    games: gs,
                };
            });

            // Build pieces for piecewise visualMap (one color per position, sorted)
            var posSet = new Set(calData.map(d => d.position));
            var positionsUsed = this.positions.filter(p => posSet.has(p));
            var pieces = positionsUsed.map(p => ({
                value: positionsUsed.indexOf(p),
                label: p,
                color: posColorMap[p],
            }));

            // Data for heatmap: [date, positionIndex]
            var seriesData = calData.map(d => [d.date, positionsUsed.indexOf(d.position)]);

            var year = this.season;
            var chart = this.ecInit('chart-calendar', 180);
            if (!chart) return;
            chart.setOption({
                tooltip: {
                    formatter: function(params) {
                        var entry = calData.find(d => d.date === params.data[0]);
                        if (!entry) return '';
                        var d = entry.date;
                        var deFmt = d.slice(8,10) + '.' + d.slice(5,7) + '.' + d.slice(0,4);
                        var lines = ['<b>' + deFmt + '</b> (' + entry.count + ' Spiel' + (entry.count > 1 ? 'e' : '') + ')'];
                        entry.games.forEach(g => {
                            var color = posColorMap[g.position] || '#888';
                            lines.push('<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + color + ';margin-right:4px"></span>'
                                + g.position + ' — ' + (g.home || '?') + ' vs ' + (g.away || '?')
                                + (g.venue ? ' <span style="color:#999">(' + g.venue + ')</span>' : ''));
                        });
                        return lines.join('<br>');
                    },
                },
                visualMap: {
                    type: 'piecewise',
                    pieces: pieces,
                    orient: 'horizontal',
                    left: 'center', bottom: 0,
                    textStyle: {color: dark ? '#D8DEE9' : '#2E3440', fontSize: 10},
                },
                calendar: {
                    range: year,
                    top: 30,
                    left: 40,
                    right: 10,
                    bottom: 30,
                    cellSize: ['auto', 15],
                    itemStyle: {
                        borderColor: dark ? '#2E3440' : '#ECEFF4',
                        borderWidth: 1,
                        color: dark ? '#3B4252' : '#E5E9F0',
                    },
                    splitLine: {lineStyle: {color: dark ? '#4C566A' : '#D8DEE9'}},
                    yearLabel: {show: false},
                    dayLabel: {
                        firstDay: 1,
                        nameMap: ['So', 'Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa'],
                        color: dark ? '#D8DEE9' : '#2E3440',
                        fontSize: 10,
                    },
                    monthLabel: {
                        nameMap: ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'],
                        color: dark ? '#D8DEE9' : '#2E3440',
                        fontSize: 10,
                    },
                },
                series: [{
                    type: 'heatmap',
                    coordinateSystem: 'calendar',
                    data: seriesData,
                }],
            });
        },

        renderFeeBar(chartId, valueFn) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var allMonths = ['Jan','Feb','Mär','Apr','Mai','Jun','Jul','Aug','Sep','Okt','Nov','Dez'];
            var f = this.filtered;
            var data = {};
            var monthsPresent = new Set();
            f.forEach(g => {
                var m = parseInt(g.month) - 1;
                data[g.position] = data[g.position] || {};
                data[g.position][m] = (data[g.position][m] || 0) + valueFn(g);
                monthsPresent.add(m);
            });
            var sortedMonths = [...monthsPresent].sort((a, b) => a - b);
            var monthLabels = sortedMonths.map(m => allMonths[m] + ' ' + this.season);
            var posWithData = this.positions.filter(p => data[p]);

            var totals = sortedMonths.map(m => {
                var t = 0;
                posWithData.forEach(p => { t += (data[p] && data[p][m]) || 0; });
                return t;
            });

            var chart = this.ecInit(chartId);
            if (!chart) return;
            var series = posWithData.map((p, i) => ({
                name: p,
                type: 'bar',
                stack: 'total',
                data: sortedMonths.map(m => data[p][m] || 0),
                itemStyle: {color: colors[i % colors.length]},
            }));
            if (series.length > 0) {
                series[series.length - 1].label = {
                    show: true,
                    position: 'right',
                    fontSize: 10,
                    formatter: function(params) {
                        return Math.round(totals[params.dataIndex]) + ' €';
                    },
                };
            }
            chart.setOption({
                tooltip: {
                    trigger: 'axis', axisPointer: {type: 'shadow'},
                    valueFormatter: v => Math.round(v) + ' €',
                },
                legend: {data: posWithData, bottom: 0, textStyle: {fontSize: 10}},
                grid: {left: 10, right: 60, top: 10, bottom: 30, containLabel: true},
                xAxis: {type: 'value', axisLabel: {formatter: v => Math.round(v) + ' €'}},
                yAxis: {type: 'category', data: monthLabels},
                series: series,
            });
        },



        renderVenueTop(chartId, games) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var counts = {};
            games.forEach(g => {
                var venue = g.venue || '';
                if (!venue) return;
                var city = venue.split(', ')[0];
                counts[city] = (counts[city] || 0) + 1;
            });
            var sorted = Object.entries(counts).sort((a, b) => a[1] - b[1]);
            var top = sorted.slice(-10);
            var labels = top.map(e => e[0]);
            var values = top.map(e => e[1]);

            var chart = this.ecInit(chartId, 500);
            if (!chart) return;
            chart.setOption({
                tooltip: {trigger: 'axis', axisPointer: {type: 'shadow'}},
                grid: {left: 10, right: 20, top: 10, bottom: 10, containLabel: true},
                xAxis: {type: 'value'},
                yAxis: {type: 'category', data: labels},
                series: [{
                    type: 'bar',
                    data: values.map((v, i) => ({value: v, itemStyle: {color: colors[i % colors.length]}})),
                    label: {show: true, position: 'right', fontSize: 11},
                }],
            });
        },

        _majorCities: [
            // Landeshauptstädte
            {name: 'Berlin', lon: 13.405, lat: 52.520},
            {name: 'München', lon: 11.582, lat: 48.135},
            {name: 'Stuttgart', lon: 9.182, lat: 48.776},
            {name: 'Düsseldorf', lon: 6.783, lat: 51.228},
            {name: 'Wiesbaden', lon: 8.240, lat: 50.083},
            {name: 'Hannover', lon: 9.732, lat: 52.375},
            {name: 'Mainz', lon: 8.271, lat: 49.999},
            {name: 'Kiel', lon: 10.139, lat: 54.323},
            {name: 'Saarbrücken', lon: 6.997, lat: 49.234},
            {name: 'Dresden', lon: 13.738, lat: 51.051},
            {name: 'Magdeburg', lon: 11.632, lat: 52.131},
            {name: 'Erfurt', lon: 11.030, lat: 50.985},
            {name: 'Potsdam', lon: 13.066, lat: 52.396},
            {name: 'Schwerin', lon: 11.417, lat: 53.636},
            {name: 'Hamburg', lon: 9.993, lat: 53.551},
            {name: 'Bremen', lon: 8.807, lat: 53.075},
            // Große Metropolen
            {name: 'Köln', lon: 6.960, lat: 50.938},
            {name: 'Frankfurt', lon: 8.682, lat: 50.110},
            {name: 'Dortmund', lon: 7.466, lat: 51.514},
            {name: 'Essen', lon: 7.012, lat: 51.458},
            {name: 'Leipzig', lon: 12.387, lat: 51.340},
            {name: 'Nürnberg', lon: 11.078, lat: 49.454},
        ],

        renderMap(chartId, games) {
            if (!this._germanyMapLoaded) return;
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var venues = {};
            games.forEach(g => {
                if (!g.venue_lat || !g.venue_lon || (g.venue_lat === 0 && g.venue_lon === 0)) return;
                var key = g.venue_lat.toFixed(4) + ',' + g.venue_lon.toFixed(4);
                if (!venues[key]) {
                    venues[key] = {lat: g.venue_lat, lon: g.venue_lon, name: g.venue || '', count: 0};
                }
                venues[key].count++;
            });
            var entries = Object.values(venues);
            if (!entries.length) return;

            var maxCount = Math.max(...entries.map(e => e.count));
            var data = entries.map(e => ({
                name: e.name,
                value: [e.lon, e.lat, e.count],
            }));

            var lats = entries.map(e => e.lat);
            var lons = entries.map(e => e.lon);
            // Padding around data bounds
            var pad = 0;
            var minLon = Math.min(...lons) - pad;
            var maxLon = Math.max(...lons) + pad;
            var minLat = Math.min(...lats) - pad;
            var maxLat = Math.max(...lats) + pad;

            // Find which Bundesländer have venues using geo containPoint
            var chart = this.ecInit(chartId, 500);
            if (!chart) return;

            // First render to get the geo coordinate system
            var activeColor = dark ? '#434C5E' : '#D8DEE9';
            var inactiveColor = dark ? '#3B4252' : '#E5E9F0';
            var borderColor = dark ? '#4C566A' : '#D8DEE9';

            // Check which regions contain venues
            var activeRegions = [];
            var mapData = echarts.getMap('Germany');
            if (mapData && mapData.geoJSON) {
                var features = mapData.geoJSON.features || [];
                features.forEach(function(feature) {
                    var name = feature.properties.name;
                    var hasVenue = entries.some(function(e) {
                        // Simple bounding box check
                        if (!feature.geometry || !feature.geometry.coordinates) return false;
                        var coords = feature.geometry.coordinates;
                        var flat = coords.flat(3);
                        var lons = [], lats = [];
                        for (var i = 0; i < flat.length; i += 2) {
                            lons.push(flat[i]);
                            lats.push(flat[i+1]);
                        }
                        var minLon = Math.min.apply(null, lons), maxLon = Math.max.apply(null, lons);
                        var minLat = Math.min.apply(null, lats), maxLat = Math.max.apply(null, lats);
                        return e.lon >= minLon && e.lon <= maxLon && e.lat >= minLat && e.lat <= maxLat;
                    });
                    if (hasVenue) {
                        activeRegions.push({
                            name: name,
                            itemStyle: {areaColor: activeColor},
                        });
                    }
                });
            }

            chart.setOption({
                geo: {
                    boundingCoords: [[minLon, maxLat], [maxLon, minLat]],
                    roam: true,
                    map: 'Germany',
                    itemStyle: {
                        areaColor: inactiveColor,
                        borderColor: borderColor,
                    },
                    emphasis: {
                        itemStyle: {
                            areaColor: dark ? '#4C566A' : '#D8DEE9',
                        },
                    },
                    regions: activeRegions,
                },
                toolbox: {
                    show: true,
                    right: 10, top: 10,
                    feature: {
                        restore: {title: 'Zurücksetzen'},
                    },
                    iconStyle: {borderColor: dark ? '#D8DEE9' : '#4C566A'},
                },
                tooltip: {
                    formatter: function(params) {
                        if (params.seriesType === 'scatter') {
                            return params.name + ' (' + params.value[2] + ' Spiele)';
                        }
                        return params.name;
                    },
                },
                series: [{
                    name: 'Venues',
                    type: 'scatter',
                    coordinateSystem: 'geo',
                    data: data,
                    symbolSize: function(val) {
                        return Math.max(8, Math.sqrt(val[2] / maxCount) * 40);
                    },
                    itemStyle: {color: '#5E81AC', opacity: 0.7},
                    label: {
                        show: true,
                        formatter: '{b}',
                        position: 'right',
                        fontSize: 10,
                        color: dark ? '#D8DEE9' : '#2E3440',
                    },
                    z: 10,
                }, {
                    name: 'Städte',
                    type: 'scatter',
                    coordinateSystem: 'geo',
                    data: this._majorCities.map(c => ({
                        name: c.name,
                        value: [c.lon, c.lat],
                    })),
                    symbol: 'circle',
                    symbolSize: 4,
                    itemStyle: {color: dark ? '#616E88' : '#9AA2AE', opacity: 0.9},
                    label: {
                        show: true,
                        formatter: '{b}',
                        position: 'bottom',
                        fontSize: 10,
                        color: dark ? '#616E88' : '#9AA2AE',
                    },
                    silent: true,
                    z: 1,
                }],
            });
        },


    }));
});
