<template>
    <div>{{chart.name}}</div>
</template>

<script>
    export default {
        props: {
            chart: {
                type: Object,
                required: true
            }
        }
    }
</script>