<template>
  <div class="user-settings">
    <a class="settings-button" title="Dark theme" href="#" @click.prevent="toggleDarkTheme">
      <DarkThemeIcon class="settings-icon" />
    </a>
  </div>
</template>

<script>
import DarkThemeIcon from "@theme/components/settings/DarkThemeIcon.vue";
import darkThemeHandler from "@theme/mixins/darkThemeHandler.js";

export default {
  name: "UserSettings",
  mixins: [darkThemeHandler],
  components: {
    DarkThemeIcon
  }
};
</script>

<style lang="stylus">
@import '../../styles/variables.styl';

.user-settings {
  position: relative;
  margin-right: 1em;

  .settings-button {
    color: $textColor;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;

    .settings-icon {
      width: 20px;
    }
  }
}

.theme-dark {
  .user-settings-menu {
    background-color: $darkPrimaryBg;
    border-color: $darkBorderColor;

    &::before {
      border-bottom-color: $darkBorderColor;
    }
  }
}

@media (max-width: $MQMobile) {
  .user-settings {
    margin-right: 0;
  }
}
</style>
