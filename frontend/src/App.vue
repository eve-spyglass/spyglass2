<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawer" clipped temporary app>
      <v-list dense>
        <v-list-item>
          <v-list-item-icon>
            <v-icon>mdi-settings</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Spyglass Configuration</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-divider></v-divider>
      </v-list>

      <settings></settings>

    </v-navigation-drawer>
    <v-app-bar app fixed clipped-left dense>
      <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>Spyglass 2</v-toolbar-title>
      <ErrorIndicator></ErrorIndicator>
    </v-app-bar>
    <v-main style="background: #ffffff">
      <v-container fluid class="px-0">
        <v-row no-gutters>
          <v-col cols="12" sm="8" md="9">
            <v-layout justify-center align-center class="px-0">
              <spyglass-map></spyglass-map>
            </v-layout>
          </v-col>
          <v-col cols="12" sm="4" md="3">
            <IntelList></IntelList>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import SpyglassMap from "./components/SpyglassMap.vue";
import Settings from "./components/Settings.vue"
import ErrorIndicator from "./components/ErrorIndicator";
import IntelList from "@/components/IntelList";

export default {
  data: () => ({
    drawer: false,
  }),
  components: {
    SpyglassMap,
    Settings,
    ErrorIndicator,
    IntelList
  },
  props: {
    source: String,
  },
};
</script>

<style>
.logo {
  width: 16em;
}
</style>
