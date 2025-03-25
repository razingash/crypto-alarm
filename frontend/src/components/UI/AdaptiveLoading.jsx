import React from 'react';

const AdaptiveLoading = () => {
    return (
        <svg className={"loading__spinner"} viewBox="0 0 50 50">
            <defs>
                <linearGradient id="gradient" gradientUnits="userSpaceOnUse" x1="0" y1="0" x2="100%" y2="0">
                    <stop offset="0%" style={{stopColor: "#2e6eff"}} />
                    <stop offset="50%" style={{stopColor: "#1f67ff"}} />
                    <stop offset="100%" style={{stopColor: "#2200ff"}} />
                </linearGradient>
            </defs>
            <circle className="spinner__stroke" cx="25" cy="25" r="20"></circle>
        </svg>
    );
};

export default AdaptiveLoading;