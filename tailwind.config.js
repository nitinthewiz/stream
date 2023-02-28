/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: ["./assets/**/*.js", "./templates/**/*.tmpl"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
