<template>
  <div id="app">
    <Donghua/>
    <LoginChart/>
  </div>
</template>

<script>
  import Donghua from './components/Donghua.vue'
  import LoginChart from './components/Chart.vue'
  export default {
    name: 'App',
    components: {
      Donghua,
      LoginChart,
    }
  }
</script>

<style>
  #app {
    margin: 0px;
    padding: 0px;
  }
  body{
    margin: 0px;
    padding: 0px;
    overflow: hidden;
    position: fixed;
    top: 0px;
    left: 0px;
    font-size: 12px;
  }
</style>
