import "../styles/strategy.css"
import FormulaEditor from "../components/FormulasEditor/FormulaEditor";

const NewStrategy = () => {

    return (
        <div className={"section__main"}>
            <div className={"field__new_formula"}>
                <div className={"area__new_formula"}>
                    {<FormulaEditor/>}
                    <button className="formula__apply_button">apply</button>
                </div>
            </div>
        </div>
    );
};

export default NewStrategy;