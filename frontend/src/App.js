import AppRouter from "./components/AppRouter";
import Header from "./components/UI/Header";
import {BrowserRouter} from "react-router-dom";
import {AuthProvider} from "./hooks/context/useAuth";
import "./styles/index.css"
import {NotificationProvider} from "./utils/store";

function App() {
    return (
        <AuthProvider>
            <NotificationProvider>
                <BrowserRouter>
                    <Header/>
                    <AppRouter/>
                </BrowserRouter>
            </NotificationProvider>
        </AuthProvider>

    );
}

export default App;
