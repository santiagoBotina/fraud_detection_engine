import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
  },
  define: {
    'import.meta.env.VITE_TRANSACTION_API_URL': JSON.stringify(
      process.env.VITE_TRANSACTION_API_URL || 'http://localhost:3000'
    ),
    'import.meta.env.VITE_DECISION_API_URL': JSON.stringify(
      process.env.VITE_DECISION_API_URL || 'http://localhost:3001'
    ),
    'import.meta.env.VITE_FRAUD_SCORE_API_URL': JSON.stringify(
      process.env.VITE_FRAUD_SCORE_API_URL || 'http://localhost:3002'
    ),
  },
})
