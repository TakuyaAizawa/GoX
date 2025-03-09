import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import PrivateRoute from './components/common/PrivateRoute';

function App() {
  return (
    <Router>
      <Routes>
        {/* 認証不要のルート */}
        <Route path="/login" element={<LoginPage />} />
        
        {/* 認証が必要なルート */}
        <Route element={<PrivateRoute />}>
          <Route path="/" element={<div className="p-4">ここにホームページが表示されます</div>} />
          <Route path="/profile/:username" element={<div className="p-4">ここにプロフィールページが表示されます</div>} />
          <Route path="/notifications" element={<div className="p-4">ここに通知ページが表示されます</div>} />
        </Route>
        
        {/* 404ページ */}
        <Route path="*" element={<div className="p-4">ページが見つかりません</div>} />
      </Routes>
    </Router>
  );
}

export default App;
