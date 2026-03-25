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
        chartLib: localStorage.getItem('db_chartLib') || 'plotly',

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
                        if (this.chartLib === 'echarts') {
                            this.renderMapEC('chart-map', this.filtered);
                            if (this.view === 'overview') {
                                this.renderMapEC('chart-overview-map', this.overviewFiltered);
                            }
                        }
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
                    this.$nextTick(() => this.renderCharts());
                }
            });
            this.$watch('chartLib', (v) => {
                localStorage.setItem('db_chartLib', v);
                this.disposeECharts();
                this.$nextTick(() => {
                    this.renderCharts();
                    if (this.view === 'overview') this.renderOverviewCharts();
                });
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
            var ec = this.chartLib === 'echarts';
            ec ? this.renderPositionChartEC() : this.renderPositionChart();
            ec ? this.renderMonthlyChartEC() : this.renderMonthlyChart();
            ec ? this.renderLeagueChartEC() : this.renderLeagueChart();
            if (ec) {
                this.renderFeeBarEC('chart-fee-total', g => g.fee + g.travel);
                this.renderFeeBarEC('chart-fee-base', g => g.fee);
                this.renderFeeBarEC('chart-fee-travel', g => g.travel);
            } else {
                this.renderFeeBar('chart-fee-total', g => g.fee + g.travel);
                this.renderFeeBar('chart-fee-base', g => g.fee);
                this.renderFeeBar('chart-fee-travel', g => g.travel);
            }
            ec ? this.renderVenueTopEC('chart-venues-top', this.filtered) : this.renderVenueTop('chart-venues-top', this.filtered);
            ec ? this.renderMapEC('chart-map', this.filtered) : this.renderMap('chart-map', this.filtered);
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

        baseLayout(height, extra) {
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var fg = getComputedStyle(document.body).color;
            var gc = dark ? '#3B4252' : '#D8DEE9';
            var layout = {
                height: height,
                margin: {t: 10, b: 30, l: 40, r: 10},
                paper_bgcolor: 'rgba(0,0,0,0)',
                plot_bgcolor: 'rgba(0,0,0,0)',
                font: {color: fg, size: 12},
                xaxis: {gridcolor: gc},
                yaxis: {gridcolor: gc},
            };
            return Object.assign(layout, extra || {});
        },

        stackAnnotations(categories, totals, horizontal, fmt, angle) {
            var fg = getComputedStyle(document.body).color;
            return categories.map((cat, i) => {
                var val = totals[i] || 0;
                if (val === 0) return null;
                var label = fmt ? fmt(val) : String(val);
                var a = horizontal ? {
                    x: val, y: cat, text: label,
                    xanchor: 'left', yanchor: 'middle',
                    xshift: 5, showarrow: false,
                    font: {size: 11, color: fg},
                } : {
                    x: cat, y: val, text: label,
                    xanchor: 'center', yanchor: 'bottom',
                    yshift: 3, showarrow: false,
                    font: {size: 11, color: fg},
                };
                if (angle) a.textangle = angle;
                return a;
            }).filter(a => a);
        },

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
            Plotly.newPlot('chart-positions', [{
                x: values, y: labels, type: 'bar', orientation: 'h',
                marker: {color: labels.map((_, i) => colors[i % colors.length])},
                text: values, textposition: 'auto',
            }], this.baseLayout(350, {
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array', categoryarray: labels,
                    ticksuffix: '  ',
                },
            }), {responsive: true, displayModeBar: false});
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
            var traces = this.positions
                .filter(p => raw[p])
                .map((p, idx) => ({
                    x: monthLabels,
                    y: months.map(m => raw[p][m] || 0),
                    type: 'bar', name: p,
                    marker: {color: colors[idx % colors.length]},
                }));
            var totals = months.map(m => {
                var t = 0;
                this.positions.forEach(p => {
                    t += (raw[p] && raw[p][m]) || 0;
                });
                return t;
            });
            Plotly.newPlot('chart-monthly', traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                margin: {t: 20, b: 30, l: 30, r: 10},
                annotations: this.stackAnnotations(monthLabels, totals, false),
            }), {responsive: true, displayModeBar: false});
        },

        // --- ECharts versions ---

        renderPositionChartEC() {
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

        renderMonthlyChartEC() {
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

        renderLeagueChartEC() {
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

        renderTreemapEC(chartId, games) {
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
            var traces = posOrder.map((p, i) => ({
                x: leagues.map(lg =>
                    (leagueData[lg] && leagueData[lg][p]) || 0
                ),
                y: leagues,
                type: 'bar', orientation: 'h', name: p,
                marker: {color: colors[i % colors.length]},
            }));
            var leagueTotals = leagues.map(lg => {
                return Object.values(leagueData[lg])
                    .reduce((a, b) => a + b, 0);
            });
            Plotly.newPlot('chart-leagues', traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array', categoryarray: leagues,
                    ticksuffix: '  ',
                },
                margin: {t: 10, b: 30, l: 100, r: 30},
                annotations: this.stackAnnotations(leagues, leagueTotals, true),
            }), {responsive: true, displayModeBar: false});
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
            this.$nextTick(() => this.renderOverviewCharts());
        },

        yearLabels(years) {
            return years.map(y => '\u200b' + y);
        },

        yearTickVals(xl, years) {
            if (years.length <= 10) return xl;
            return xl.filter((_, i) => parseInt(years[i]) % 5 === 0);
        },

        renderOverviewCharts() {
            if (!this.overviewLoaded) return;
            this.renderOverviewGamesPerYear();
            this.renderOverviewPositionTrend();
            this.renderOverviewPositionPie();
            this.renderOverviewFeePerYear();
            this.renderOverviewAvgPerGame();
            this.renderOverviewKmPerYear();
            this.renderOverviewSankey();
            this.renderOverviewLeagues();
            this.renderVenueTop('chart-overview-venues-top', this.overviewFiltered);
            this.renderMap('chart-overview-map', this.overviewFiltered);
        },

        renderOverviewGamesPerYear() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var traces = this.overviewPositions
                .filter(p => years.some(y =>
                    (byYear[y].by_position[p] || 0) > 0
                ))
                .map((p, i) => ({
                    x: xl,
                    y: years.map(y =>
                        byYear[y].by_position[p] || 0
                    ),
                    type: 'bar', name: p,
                    marker: {color: colors[i % colors.length]},
                }));
            var totals = years.map(y => byYear[y].count);
            Plotly.newPlot('chart-overview-games', traces,
                this.baseLayout(350, {
                    barmode: 'stack',
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 30, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                    annotations: this.stackAnnotations(
                        xl, totals, false
                    ),
                }),
                {responsive: true, displayModeBar: false});
        },

        renderOverviewPositionTrend() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var traces = this.overviewPositions
                .filter(p => years.some(y =>
                    (byYear[y].by_position[p] || 0) > 0
                ))
                .map((p, i) => ({
                    x: xl,
                    y: years.map(y =>
                        byYear[y].by_position[p] || 0
                    ),
                    type: 'scatter', mode: 'lines+markers',
                    name: p,
                    line: {
                        color: colors[i % colors.length], width: 2,
                    },
                    marker: {size: 5},
                }));
            Plotly.newPlot('chart-overview-trend', traces,
                this.baseLayout(350, {
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 30, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                }),
                {responsive: true, displayModeBar: false});
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
            var sorted = Object.entries(totals)
                .sort((a, b) => b[1] - a[1]);
            Plotly.newPlot('chart-overview-pie', [{
                labels: sorted.map(e => e[0]),
                values: sorted.map(e => e[1]),
                type: 'pie', hole: 0.4,
                marker: {colors: colors},
                textinfo: 'label+percent',
                textposition: 'outside',
                textfont: {size: 11},
                domain: {x: [0.1, 0.9], y: [0.1, 0.9]},
            }], this.baseLayout(350, {
                showlegend: false,
                margin: {t: 30, b: 30, l: 30, r: 30},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewFeePerYear() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var feeTrace = {
                x: xl,
                y: years.map(y =>
                    byYear[y] ? byYear[y].fee : 0
                ),
                type: 'bar', name: 'Pauschale',
                marker: {color: '#5E81AC'},
            };
            var travelTrace = {
                x: xl,
                y: years.map(y =>
                    byYear[y] ? byYear[y].travel : 0
                ),
                type: 'bar', name: 'Fahrtkosten',
                marker: {color: '#88C0D0'},
            };
            var totals = years.map(y => {
                var d = byYear[y];
                return d ? d.fee + d.travel : 0;
            });
            var fmtEur = v => Math.round(v) + ' €';
            var traces = [feeTrace, travelTrace];
            Plotly.newPlot('chart-overview-fee', traces,
                this.baseLayout(350, {
                    barmode: 'stack',
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 50, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                    annotations: this.stackAnnotations(
                        xl, totals, false, fmtEur, -45
                    ),
                }),
                {responsive: true, displayModeBar: false});
        },

        renderOverviewAvgPerGame() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var avgs = years.map(y => {
                var d = byYear[y];
                if (!d || d.count === 0) return 0;
                return (d.fee + d.travel) / d.count;
            });
            var fmtEur = v => v.toFixed(0) + ' €';
            Plotly.newPlot('chart-overview-avg', [{
                x: xl, y: avgs, type: 'bar',
                marker: {color: '#A3BE8C'},
                text: avgs.map(fmtEur), textposition: 'auto',
            }], this.baseLayout(350, {
                margin: {t: 20, b: 30, l: 50, r: 10},
                xaxis: {tickvals: this.yearTickVals(xl, years)},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewKmPerYear() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var kms = years.map(y =>
                byYear[y] ? byYear[y].km : 0
            );
            Plotly.newPlot('chart-overview-km', [{
                x: xl, y: kms, type: 'bar',
                marker: {color: '#81A1C1'},
                text: kms.map(v => v.toLocaleString('de-DE')),
                textposition: 'auto',
            }], this.baseLayout(350, {
                margin: {t: 20, b: 30, l: 50, r: 10},
                xaxis: {tickvals: this.yearTickVals(xl, years)},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewSankey() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.overviewFiltered;
            if (!games.length) return;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var positions = this.overviewPositions.filter(
                p => games.some(g => g.position === p)
            );
            var counts = {};
            games.forEach(g => {
                var key = g.position + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });
            var z = positions.map(p =>
                years.map(y => counts[p + '|' + y] || 0)
            );
            var annotations = [];
            positions.forEach((p, pi) => {
                years.forEach((y, yi) => {
                    var val = counts[p + '|' + y] || 0;
                    if (val > 0) {
                        annotations.push({
                            x: xl[yi], y: p,
                            text: '' + val,
                            showarrow: false,
                            font: {color: val > 8 ? '#2E3440' : '#ECEFF4', size: 11},
                        });
                    }
                });
            });
            Plotly.newPlot('chart-overview-sankey', [{
                x: xl, y: positions, z: z,
                type: 'heatmap',
                colorscale: [
                    [0, '#3B4252'],
                    [0.3, '#5E81AC'],
                    [0.6, '#88C0D0'],
                    [0.8, '#D08770'],
                    [1, '#BF616A'],
                ],
                showscale: false,
                hoverongaps: false,
                xgap: 2,
                ygap: 2,
            }], this.baseLayout(Math.max(300, positions.length * 45), {
                margin: {t: 10, b: 30, l: 40, r: 10},
                xaxis: {tickvals: xl, side: 'bottom'},
                yaxis: {autorange: 'reversed'},
                annotations: annotations,
            }), {responsive: true, displayModeBar: false});
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
            var xl = this.yearLabels(years);
            var leagueTotals = {};
            games.forEach(g => {
                leagueTotals[g.league] = (leagueTotals[g.league] || 0) + 1;
            });
            var leagues = Object.keys(leagueTotals)
                .sort((a, b) => leagueTotals[b] - leagueTotals[a]);
            var counts = {};
            games.forEach(g => {
                var key = g.league + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });
            var traces = years.map((y, i) => ({
                x: leagues,
                y: leagues.map(lg => counts[lg + '|' + y] || 0),
                type: 'bar',
                name: y,
                marker: {color: colors[i % colors.length]},
            }));
            Plotly.newPlot('chart-overview-leagues', traces,
                this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                margin: {t: 10, b: 80, l: 40, r: 10},
                xaxis: {tickangle: -45},
            }), {responsive: true, displayModeBar: false});
        },

        renderFeeBarEC(chartId, valueFn) {
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

        renderFeeBar(chartId, valueFn) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var months = ['Jan','Feb','Mär','Apr','Mai','Jun','Jul','Aug','Sep','Okt','Nov','Dez'];
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
            var monthLabels = sortedMonths.map(m => months[m] + ' ' + this.season);
            var traces = this.positions
                .filter(p => data[p])
                .map((p, i) => ({
                    y: monthLabels,
                    x: sortedMonths.map(m => data[p][m] || 0),
                    type: 'bar', orientation: 'h', name: p,
                    marker: {color: colors[i % colors.length]},
                }));
            var totalArr = sortedMonths.map(m => {
                var t = 0;
                this.positions.forEach(p => {
                    t += (data[p] && data[p][m]) || 0;
                });
                return t;
            });
            var fmtEur = v => Math.round(v) + ' €';
            Plotly.newPlot(chartId, traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array',
                    categoryarray: monthLabels,
                    ticksuffix: '  ',
                },
                margin: {t: 10, b: 30, l: 90, r: 60},
                annotations: this.stackAnnotations(
                    monthLabels, totalArr, true, fmtEur
                ),
            }), {responsive: true, displayModeBar: false});
        },

        renderTreemap(chartId, games) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var fg = getComputedStyle(document.body).color;

            // Build hierarchy: Alle > position > league
            var labels = ['Alle'];
            var parents = [''];
            var values = [games.length];
            var markerColors = [dark ? '#2E3440' : '#ECEFF4'];

            var posData = {};
            games.forEach(g => {
                if (!posData[g.position]) posData[g.position] = {};
                posData[g.position][g.league] = (posData[g.position][g.league] || 0) + 1;
            });

            // Mix color towards white (light) or dark (dark theme)
            function shadeColor(hex, factor) {
                var r = parseInt(hex.slice(1,3), 16);
                var g = parseInt(hex.slice(3,5), 16);
                var b = parseInt(hex.slice(5,7), 16);
                var target = dark ? 30 : 240;
                r = Math.round(r + (target - r) * factor);
                g = Math.round(g + (target - g) * factor);
                b = Math.round(b + (target - b) * factor);
                return '#' + [r,g,b].map(c => c.toString(16).padStart(2,'0')).join('');
            }

            var posOrder = this.positions.filter(p => posData[p]);
            posOrder.forEach((pos, i) => {
                var baseColor = colors[i % colors.length];
                var total = Object.values(posData[pos]).reduce((a, b) => a + b, 0);
                labels.push(pos);
                parents.push('Alle');
                values.push(total);
                markerColors.push(baseColor);

                var leagues = Object.entries(posData[pos]).sort((a, b) => b[1] - a[1]);
                leagues.forEach(([lg, count], j) => {
                    labels.push(pos + '/' + lg);
                    parents.push(pos);
                    values.push(count);
                    var shade = 0.2 + (j / Math.max(leagues.length - 1, 1)) * 0.4;
                    markerColors.push(shadeColor(baseColor, shade));
                });
            });

            // Display text: league name for leaves, "Pos (count)" for parents
            var displayText = labels.map((l, idx) => {
                var slash = l.indexOf('/');
                if (slash >= 0) return l.substring(slash + 1);
                if (l === 'Alle') return 'Alle (' + values[idx] + ')';
                return l + ' (' + values[idx] + ')';
            });

            Plotly.newPlot(chartId, [{
                type: 'treemap',
                ids: labels,
                labels: displayText,
                parents: parents,
                values: values,
                marker: {
                    colors: markerColors,
                    line: {color: dark ? '#2E3440' : '#ECEFF4', width: 2},
                },
                textinfo: 'label+value',
                textfont: {size: 12},
                branchvalues: 'total',
                pathbar: {visible: true, textfont: {size: 11}},
                tiling: {packing: 'squarify', pad: 3},
                maxdepth: 3,
                level: 'Alle',
            }], {
                height: 350,
                margin: {t: 25, b: 0, l: 0, r: 0},
                paper_bgcolor: 'rgba(0,0,0,0)',
                font: {color: fg},
            }, {responsive: true, displayModeBar: false});
        },

        renderVenueTopEC(chartId, games) {
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

        renderMapEC(chartId, games) {
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
                }],
            });
        },

        renderVenueTop(chartId, games) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var counts = {};
            games.forEach(g => {
                // Extract city from venue (format: "Stadium, City" or just "City")
                var venue = g.venue || '';
                if (!venue) return;
                var parts = venue.split(', ');
                var city = parts[0];
                counts[city] = (counts[city] || 0) + 1;
            });
            var sorted = Object.entries(counts)
                .sort((a, b) => a[1] - b[1]);
            var top = sorted.slice(-10);
            var labels = top.map(e => e[0]);
            var values = top.map(e => e[1]);

            Plotly.newPlot(chartId, [{
                x: values, y: labels, type: 'bar', orientation: 'h',
                marker: {color: labels.map((_, i) => colors[i % colors.length])},
                text: values, textposition: 'auto',
            }], this.baseLayout(500, {
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array', categoryarray: labels,
                    ticksuffix: '  ',
                },
                margin: {t: 10, b: 30, l: 120, r: 10},
            }), {responsive: true, displayModeBar: false});
        },

        renderMap(chartId, games) {
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            // Aggregate games by venue (lat/lon)
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
            var trace = {
                type: 'scattermapbox',
                lat: entries.map(e => e.lat),
                lon: entries.map(e => e.lon),
                text: entries.map(e => e.name + ' (' + e.count + ')'),
                marker: {
                    size: entries.map(e => Math.max(8, Math.sqrt(e.count / maxCount) * 40)),
                    color: '#5E81AC',
                    opacity: 0.7,
                },
                hoverinfo: 'text',
            };

            var lats = entries.map(e => e.lat);
            var lons = entries.map(e => e.lon);
            var cLat = (Math.min(...lats) + Math.max(...lats)) / 2;
            var cLon = (Math.min(...lons) + Math.max(...lons)) / 2;
            var latSpan = Math.max(...lats) - Math.min(...lats) || 1;
            var lonSpan = Math.max(...lons) - Math.min(...lons) || 1;
            var span = Math.max(latSpan, lonSpan);
            // zoom: ~6 for Germany-wide, ~8 for regional, ~10 for local
            var zoom = Math.round(Math.log2(180 / span));

            Plotly.newPlot(chartId, [trace], {
                mapbox: {
                    style: dark ? 'carto-darkmatter' : 'open-street-map',
                    center: {lat: cLat, lon: cLon},
                    zoom: zoom,
                },
                height: 500,
                margin: {t: 0, b: 0, l: 0, r: 0},
                paper_bgcolor: 'rgba(0,0,0,0)',
            }, {responsive: true, displayModeBar: false, scrollZoom: true});
        },
    }));
});
