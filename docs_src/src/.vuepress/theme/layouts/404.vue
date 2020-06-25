<template>
  <ParentLayout>
    <template slot="page-top">
      <div class="theme-default-content">
        <h1>404</h1>

        <blockquote>The page you are looking for doesn't exist.</blockquote>

        <RouterLink to="/">Go back home</RouterLink>
      </div>
    </template>
  </ParentLayout>
</template>

<script>
import ParentLayout from "@parent-theme/layouts/Layout.vue";

export default {
  components: {
    ParentLayout
  }
};
</script>

<style src="../styles/variables.styl" lang="stylus"></style>
<style src="../styles/index.styl" lang="stylus"></style>
<style src="../styles/dark.styl" lang="stylus"></style>

<style lang="stylus">
#app {
  min-height: 100vh;
}
</style>