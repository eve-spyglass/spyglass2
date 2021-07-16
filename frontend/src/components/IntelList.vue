<template>
  <v-card class="scroll" height="90vh">
    <v-card-text >
      <div id="example-1">
        <v-card v-for="item in message.slice().reverse()" :key="item">
          <v-card-title class="pa-0 ma-0" >{{JSON.parse(item).message}}</v-card-title>
          <v-card-subtitle class="pa-0 ma-0" >{{JSON.parse(item).reporter}} <span class="float-right">{{ JSON.parse(item).source }}</span></v-card-subtitle>
        </v-card>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import Wails from "@wailsapp/runtime";

export default {
  data() {
    return {
      message: [],
    };
  },
  methods: {
    getMessage: function () {
      var self = this;
      window.backend.UserInterface.GetIntelMessages().then((result) => {
        self.message = result;
      });
    },
  },
  mounted: function () {
    this.getMessage();
    Wails.Events.On("ui_update", (nothing) => {
      if (nothing) {
        this.getMessage();
      } else {
        this.getMessage();
      }
    });
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.scroll{
  overflow-y: scroll;
}
</style>
