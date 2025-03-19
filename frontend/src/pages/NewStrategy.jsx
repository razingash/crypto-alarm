import "../styles/strategy.css"
import Keyboard from "../components/Keyboard";

const NewStrategy = () => {

    return (
        <div className={"section__main"}>
            <div className={"field__new_formula"}>
                <div className={"area__new_formula"}>
                    {<Keyboard/>}
                    <button className="formula__apply_button">apply</button>
                </div>
            </div>
        </div>
    );
};

export default NewStrategy;