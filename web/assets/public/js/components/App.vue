<template>
    <Chart v-for="chart in charts" :chart="chart"></chart>
</template>

<script>
    import "./Chart"

    export default {
        comments: {
            Chart
        },
        data() {
            return {
                charts: [
                    {
                        name: "sensor1"
                    },
                    {
                        name: "sensor2"
                    }
                ]
            }
        }
    }
</script>