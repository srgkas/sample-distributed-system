Vue.component('sensorchart', {
    extends: VueChartJs.Line,
    props: {
        def: {
            type: Object,
            required: true
        }
    },
    data() {
        let now = new Date();

        return {}
    },
    mounted() {
        this.render()
    },

    methods: {
        render() {
            this.renderChart({
                labels: [new Date()],
                datasets: [
                    {
                        label: this.def.name,
                        borderColor: 'rgba(50,50,255,0.7)',
                        backgroundColor: 'rgba(0, 0, 0, 0)',
                        data: this.def.values,
                        fill: false,
                        type: 'line',
                        borderWidth: 2
                    }
                ],
            }, {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    xAxes: [
                        {
                            scaleLabel: {
                                display: true,
                                labelString: 'Time ( UTC )'
                            },
                            type: 'time',
                            // bounds: "data",
                            distribution: "series",
                            time: {
                                tooltipFormat: "hh:mm:ss",
                                displayFormats: {
                                    hour: 'MMM D, hh:mm:ss'
                                }
                            },
                            ticks: {
                                source: 'data',
                            }
                        }
                    ],
                }
            })
        },
        updateChart() {
            this.$data._chart.update()
        }
    },
    watch: {
        "def.values": function (oldValue, newValue) {
            this.updateChart()
        }
    }
})

var app = new Vue({
    el: "#app",
    data: {
        name: "Test",
        sensors: [],
        socket: null,
    },
    template: '<div><sensorchart v-for="sensor in sensors" :def="sensor"></sensorchart></div>',
    mounted() {
        this.initSocket()
    },
    methods: {
        initSocket() {
            if (this.socket === null) {
                this.socket = new WebSocket("ws://localhost:3000/ws")
            }

            this.socket.addEventListener("message", (e) => {
                let msg = JSON.parse(e.data)
                console.log("msg received", msg)

                switch (msg.type) {
                    case "source":
                        if (!this.sensors.some(function (s) {
                            return s.name === msg.data.name;
                        })) {
                            this.sensors.push({
                                name: msg.data.name,
                                values: [
                                    {
                                        t: msg.data.timestamp,
                                        y: msg.data.value,
                                    }
                                ]
                            })
                        }
                        break
                    case "reading":
                        let sensorData = msg.data
                        let sensor = this.sensors.filter((s) => s.name === sensorData.name)[0]
                        sensor.values.push(
                            {
                                t: sensorData.timestamp,
                                y: sensorData.value,
                            })
                        break
                    default:
                        console.log("wrong type of message")
                }
            })

            this.socket.addEventListener("open", (e) => {
                this.socket.send(JSON.stringify({
                    type: 'discover'
                }))
            })
        }
    },
})
