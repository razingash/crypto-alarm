import AppRouter from "./components/AppRouter";
import Header from "./components/UI/Header";
import {BrowserRouter} from "react-router-dom";
import {AuthProvider} from "./hooks/context/useAuth";
import "./styles/index.css"
import {StoreProvider} from "./utils/store";

function App() {
    return (
        <StoreProvider>
            <AuthProvider>
                <BrowserRouter>
                    <Header/>
                    <AppRouter/>
                </BrowserRouter>
            </AuthProvider>
        </StoreProvider>
    );
}

export default App;
