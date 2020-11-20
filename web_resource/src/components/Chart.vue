<template>
  <div height="200" style="width:100%;height:256px;position:absolute;bottom:0;z-index: 10">
    <canvas id="myChart" ></canvas>
  </div>
</template>
<script>
import Chart from 'chart.js';
import axios from 'axios';

export default {
  data() {
    return {
      xlist:
          [],
      ylist: []
    }
  },
  methods: {
    getChartData: function() {
      var that = this
      axios.get('/login_charts').then(function(response) {
        console.info(response)
        that.xlist = response.data.x
        that.ylist = response.data.y

        that.renderCharts()
      })
    },
    renderCharts: function() {
      var ctx = document.getElementById("myChart");

      // ctx.setAttribute("style","height:200px;width:100%;position:absolute;bottom:0")
      console.info(this.xlist, this.ylist)

      var options = {
        maintainAspectRatio: false,
        spanGaps: false,
        elements: {
          line: {
            tension: 0.4
          }
        },
        plugins: {
          filler: {
            propagate: false
          }
        },
        scales: {
          xAxes: [{
            ticks: {
              autoSkip: true,
              maxRotation: 0,
              display: true,
            }
          }]
        }
      };

      new Chart(ctx, {
        type: 'line',
        data: {
          labels: this.xlist,
          datasets: [{
            backgroundColor: 'rgba(255, 99, 132, 0.5)',
            borderColor: 'rgba(255, 99, 132, 0.5)',
            data: this.ylist,
            label: '',
            fill: 'start',

          }]
        },
        options: Chart.helpers.merge(options, {
          title: {
            text: '用户上线全天时间分布图',
            display: true,
            position:'bottom',
          }
        })
      });
    }
  },
  beforeMount() {
    // this.getChartData()
  },
  mounted() {
    this.getChartData()
  }
}
</script>
