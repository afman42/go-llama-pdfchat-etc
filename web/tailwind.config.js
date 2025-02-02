export default {
  content: ["./index.html", "./src/**/*.{js}"],
  theme: {
    extend: {},
    screens: {
      "min-sm": { min: "300px", max: "600px" },
      sm: { min: "640px", max: "767px" },
      md: { min: "768px", max: "1023px" },
      lg: { min: "1024px", max: "1279px" },
      xl: { min: "1280px" },
    },
  },
  plugins: [],
};
