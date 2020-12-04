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
      const root = document.getElementsByTagName("html")[0]
      if (this.darkTheme) {
        root.classList.add('theme-dark');
        document.body.classList.add('theme-dark');
        return localStorage.setItem('dark-theme', true);
      }

      document.body.classList.remove('theme-dark');
      localStorage.setItem('dark-theme', false);
    }
  },
};
