<template>
  <v-container fluid class="px-0">
    <panZoom
        :options="{minZoom: 0.5, maxZoom: 5, zoomSpeed:0.5}"
    >
        <div id="svgMap" v-html="message"></div>
    </panZoom>
  </v-container>
</template>

<script>


import Wails from '@wailsapp/runtime';

  export default {
    data () {
      return {
        message: "",
      }
    },
    methods: {
      getMessage: function () {
        var self = this
        window.backend.EveMapper.GetCurrentMapSVG().then(result => {
          self.message = result
        })
      }
    },
    mounted: function() {
      this.getMessage();
      Wails.Events.On("ui_update", nothing => {
        if (nothing) {
          this.getMessage();
        } else {
          this.getMessage();
        }
        
      });
    }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  h1 {
    margin-top: 2em;
    position: relative;
    min-height: 5rem;
    width: 100%;
  }

  /*a:hover {*/
  /*  font-size: 1.7em;*/
  /*  border-color: blue;*/
  /*  background-color: blue;*/
  /*  color: white;*/
  /*  border: 3px solid white;*/
  /*  border-radius: 10px;*/
  /*  padding: 9px;*/
  /*  cursor: pointer;*/
  /*  transition: 500ms;*/
  /*}*/

  a {
    font-size: 1.7em;
    border-color: white;
    background-color: white;
    color: white;
    border: 3px solid white;
    border-radius: 10px;
    padding: 9px;
    cursor: pointer;
  }
</style>
