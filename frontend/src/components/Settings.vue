<template>
  <v-container fluid class="px-0">
     <v-row justify="center">
      <v-dialog
        v-model="dialog"
        persistent
        max-width="600px"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            
            dark
            v-bind="attrs"
            v-on="on"
            text
          >
            Settings
          </v-btn>
        </template>
        <v-card>
          <v-card-title>
            <span class="headline">Spyglass Settings</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-form
                ref="form"
                v-model="valid"
                lazy-validation
              >
                <v-select
                  v-model="region"
                  :items="regionOptions"
                  label="Map"
                  autocomplete
                  required
                ></v-select>

                <v-text-field
                  v-model="chatlogDir"
                  label="Chat Log Directory"
                  clearable
                ></v-text-field>

                <v-text-field
                  v-model="channels"
                  label="Intel Channels"
                  clearable
                ></v-text-field>

              </v-form>
              <v-card
                class="mx-auto"
                outlined
                tile
                elevation="3"
              >
                <v-card-text>
                  <h3 class="text--primary">WARNING</h3>
                  <p>Any changed settings will not come into effect until spyglass is restarted</p>
                </v-card-text>
              </v-card>
            </v-container>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn
              color="blue darken-1"
              text
              @click="cancelFunc"
            >
              Cancel
            </v-btn>
            <v-btn
              color="blue darken-1"
              text
              @click="saveFunc"
            >
              Save
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-row>
  </v-container>
</template>

<script>
  export default {
    data () {
      return {
        dialog: false,
        chatlogDir: "",
        channels: "",
        region: null,
        regionOptions: [""],
      }
    },
    mounted: function() {
      this.loadConfig();
    },
    methods: {
      loadConfig: function() {
        window.backend.Config.GetData().then(result => {
          var d = JSON.parse(result);

          this.chatlogDir = d.chatLogDirectory;
          this.region = d.selectedMap;
          this.channels = d.channels.join(";");
        })

        window.backend.EveMapper.GetAvailableMaps().then(result => {
          this.regionOptions = result;
        })
      },

      saveConfig: function() {

        var cfg = {
          selectedMap: this.region,
          chatLogDirectory: this.chatlogDir,
          channels: this.channels.split(";")
        }

        window.backend.Config.SetConfig(cfg)
      },

      cancelFunc: function() {
        this.loadConfig();
        this.dialog = false;
      },

      saveFunc: function() {
        this.saveConfig();
        this.dialog = false;
      }
    },
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
