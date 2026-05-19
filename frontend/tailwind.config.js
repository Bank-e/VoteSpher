/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50:  '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
        emerald: {
          50:  '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
        },
        amber: {
          50:  '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          700: '#b45309',
        },
        red: {
          50:  '#fef2f2',
          200: '#fecaca',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
        },
        purple: {
          50:  '#faf5ff',
          200: '#e9d5ff',
          400: '#c084fc',
          700: '#7e22ce',
        },
      },
      fontFamily: {
        sans: ['Sarabun', 'Inter', 'sans-serif'],
      },
      animation: {
        'fade-in':  'fadeIn 0.35s ease-out',
        'slide-up': 'slideUp 0.4s ease-out',
        'pop':      'pop 0.25s cubic-bezier(0.34,1.56,0.64,1)',
      },
      keyframes: {
        fadeIn:  { from: { opacity: 0 },                                          to: { opacity: 1 } },
        slideUp: { from: { opacity: 0, transform: 'translateY(14px)' },           to: { opacity: 1, transform: 'none' } },
        pop:     { from: { opacity: 0, transform: 'scale(0.9)' },                 to: { opacity: 1, transform: 'scale(1)' } },
      },
    },
  },
  plugins: [],
}
