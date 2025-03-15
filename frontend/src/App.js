import AppRouter from "./components/AppRouter";
import Header from "./components/UI/Header";
import {BrowserRouter} from "react-router-dom";
import {AuthProvider} from "./hooks/context/useAuth";

function App() {
    return (
        <AuthProvider>
            <BrowserRouter>
                <Header/>
                <AppRouter/>
            </BrowserRouter>
        </AuthProvider>
    );
}

export default App;
