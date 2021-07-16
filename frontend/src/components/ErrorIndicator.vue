<template>
  <v-row justify="center">
    <v-dialog
        v-model="dialog"
        scrollable
        max-width="300px"
    >
      <template v-slot:activator="{ on, attrs }">
        <v-btn
            color="warning"
            dark
            v-bind="attrs"
            v-on="on"
            v-show="errors"
        >
          Error List
        </v-btn>
      </template>
      <v-card>
        <v-card-title>Error List</v-card-title>
        <v-divider></v-divider>
        <v-card-text style="height: 300px">
          <ul id="example-1">
            <li v-for="item in message.slice().reverse()" :key="item">
              {{ item }}
            </li>
          </ul>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-btn
              color="blue darken-1"
              text
              @click="dialog = false"
          >
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>

import Wails from '@wailsapp/runtime';

  export default {
    data () {
      return {
        errors: false,
        dialog: false,
        message: [],
      }
    },
    methods: {
      getMessage: function () {
        var self = this
        window.backend.UserInterface.ReadErrorList().then(result => {
          self.message = result;
          if (self.message.length > 0){
            self.errors = true;
          } else {
            self.errors = false;
          }
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

</style>
