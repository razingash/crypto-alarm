import AppRouter from "./components/AppRouter";
import Header from "./components/UI/Header";
import {BrowserRouter} from "react-router-dom";
import "./styles/index.css"
import {NotificationProvider} from "./utils/store";

function App() {
    return (
        <NotificationProvider>
            <BrowserRouter>
                <Header/>
                <AppRouter/>
            </BrowserRouter>
        </NotificationProvider>
    );
}

export default App;
