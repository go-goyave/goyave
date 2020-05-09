<template>
  <component :is="$page.frontmatter.homeComponent || 'ParentHome'" />
</template>

<script>
import ParentHome from "@parent-theme/components/Home.vue";

export default {
  components: {
    ParentHome
  }
};
</script>
