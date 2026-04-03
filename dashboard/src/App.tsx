import { BrowserRouter, Routes, Route } from "react-router-dom";
import TransactionList from "./pages/TransactionList";
import TransactionDetail from "./pages/TransactionDetail";
import Navbar from "./components/Navbar";
import FadeWrapper from "./components/FadeWrapper";
import { useTheme } from "./hooks/useTheme";

function AppContent() {
  const { theme, toggleTheme } = useTheme();

  return (
    <>
      <Navbar theme={theme} onToggleTheme={toggleTheme} />
      <main style={{ paddingTop: "72px", padding: "72px 24px 48px" }}>
        <FadeWrapper>
          <Routes>
            <Route path="/" element={<TransactionList />} />
            <Route path="/transactions/:id" element={<TransactionDetail />} />
          </Routes>
        </FadeWrapper>
      </main>
    </>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}

export default App;
