<template>
  <div id="app">
    <!--<Header/>-->
    <!--<Body/>-->
    <Donghua/>
  </div>
</template>

<script>
  // import Header from './components/Header.vue'
  // import Body from './components/Body.vue'
  import Donghua from './components/Donghua.vue'

  export default {
    name: 'App',
    components: {
      // Header,
      // Body,
      Donghua
    }
  }
</script>

<style>
  #app {
    font-family: Avenir, Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-align: center;
    color: #2c3e50;
  }
</style>
