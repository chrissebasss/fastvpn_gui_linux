/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#090B10",
        foreground: "#F1F5F9",
        card: {
          DEFAULT: "#0F131C",
          foreground: "#F1F5F9",
          elevated: "#141A26",
          hover: "#1B2232",
        },
        popover: {
          DEFAULT: "#101520",
          foreground: "#F1F5F9",
        },
        primary: {
          DEFAULT: "#FF8A3D",
          foreground: "#FFFFFF",
          hover: "#FF771C",
          muted: "rgba(255, 138, 61, 0.15)",
        },
        secondary: {
          DEFAULT: "#18202F",
          foreground: "#E2E8F0",
        },
        muted: {
          DEFAULT: "#131824",
          foreground: "#94A3B8",
        },
        accent: {
          DEFAULT: "#1F293D",
          foreground: "#F8FAFC",
        },
        destructive: {
          DEFAULT: "#EF4444",
          foreground: "#FFFFFF",
          hover: "#DC2626",
        },
        border: "#20283A",
        input: "#161D2B",
        ring: "#FF8A3D",
        success: {
          DEFAULT: "#22C55E",
          foreground: "#FFFFFF",
          muted: "rgba(34, 197, 94, 0.15)",
        },
      },
      borderRadius: {
        lg: "0.5rem",
        md: "0.375rem",
        sm: "0.25rem",
      },
    },
  },
  plugins: [],
};

