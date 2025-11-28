/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        mc: {
          grass: '#7CBD37',    // Minecraft Grass Green
          dirt: '#8B572A',     // Rich earthy brown
          stone: '#7D7D7D',    // Classic stone grey
          diamond: '#2DD4BF',  // Cyan/Teal
          gold: '#FCD34D',     // Bright yellow-gold
          redstone: '#EF4444', // Vibrant red
          obsidian: '#1F1F1F', // Dark grey/black
        },
        bg: '#E5E5E5',         // Light Stone Grey
      },
      boxShadow: {
        'neo': '4px 4px 0px 0px rgba(0,0,0,1)',
        'neo-lg': '8px 8px 0px 0px rgba(0,0,0,1)',
        'neo-sm': '2px 2px 0px 0px rgba(0,0,0,1)',
      },
      borderWidth: {
        '3': '3px',
      },
      fontWeight: {
        'neo': '800',
      },
      fontFamily: {
        'mono': ['"VT323"', 'monospace'], // We'll add this font
        'sans': ['"Inter"', 'sans-serif'],
      }
    },
  },
  plugins: [],
}