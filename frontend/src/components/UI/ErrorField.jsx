import React from 'react';

const ErrorField = ({message}) => {
    return (
        <div className="field__ise">
            <div className={"ise_description"}>
                {message ? message : "Backend server is most likely offline"}
            </div>
        </div>
    );
};

export default ErrorField;