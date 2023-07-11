/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./static/html/*.html", 
    "./static/script/*.{js,ts}", 
  ],
  theme: {
    fontFamily: {
      'prompt': ['Prompt'],
      'heading': ['Boogaloo']
    },
    extend: {},
  },
  plugins: [],
}

