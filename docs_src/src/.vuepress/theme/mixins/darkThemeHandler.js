export default {
  data() {
    return {
      darkTheme: false,
    };
  },

  mounted() {
    this.darkTheme = localStorage.getItem('dark-theme') === 'true';
    this.updateTheme();
  },

  methods: {
    toggleDarkTheme() {
      this.darkTheme = !this.darkTheme;
      this.updateTheme();
    },
    updateTheme() {
      if (this.darkTheme) {
        document.body.classList.add('yuu-theme-dark');
        return localStorage.setItem('dark-theme', true);
      }

      document.body.classList.remove('yuu-theme-dark');
      localStorage.setItem('dark-theme', false);
    }
  },
};
