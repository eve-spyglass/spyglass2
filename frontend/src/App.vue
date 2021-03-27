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
    </v-app-bar>
    <v-main style="background: #ffffff">
      <v-container fluid class="px-0">
        <v-layout justify-center align-center class="px-0">
          <spyglass-map></spyglass-map>
        </v-layout>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import SpyglassMap from "./components/SpyglassMap.vue";
import Settings from "./components/Settings.vue"

export default {
  data: () => ({
    drawer: false,
  }),
  components: {
    SpyglassMap,
    Settings,
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
