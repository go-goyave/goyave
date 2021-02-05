<template>
  <div class="banner" v-if="!hide">
    <p class="custom-block warning">
      <span class="close-btn" title="Dismiss" @click="hide = true">&times;</span>
      <span>Goyave's website has moved to&nbsp;<a href="https://goyave.dev" rel="alternate">https://goyave.dev</a>. This location will soon be removed.</span>
    </p>
  </div>
</template>

<script>
export default {
  name: "MigrationBanner",
  data() {
    return {
      hide: false
    }
  }
}
</script>

<style scoped lang="stylus">
.banner
  position: fixed
  flex-direction: column
  z-index: 20;
  bottom: 0
  left: 0
  right: 0
  background-color: #fff

  .custom-block
    padding: 0.75rem 1.5rem
    border-left: none
    margin: 0
    display: flex

    & > *
      align-self: center

.theme-dark .banner
  background-color: #1a1a1a

.close-btn
  color: #2c3e50
  text-align: center
  font-size: 2rem
  padding: 0 0.5rem

  &:hover
    cursor: pointer

.theme-dark .close-btn
  color: #fff
</style>

<style lang="stylus">
.go-to-top
  z-index: 21 !important
</style>