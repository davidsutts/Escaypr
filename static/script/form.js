var loading = false;
var signup = false;
document.addEventListener("DOMContentLoaded", function () {
    var _a, _b;
    (_a = document.getElementById("pword")) === null || _a === void 0 ? void 0 : _a.addEventListener("change", toggleSubmit);
    (_b = document.getElementById("uname")) === null || _b === void 0 ? void 0 : _b.addEventListener("change", toggleSubmit);
});
document.addEventListener("keyup", function (event) {
    var button = document.querySelector("#submit");
    toggleSubmit();
    if (event.code == "Enter") {
        if (checkForm()) {
            postForm();
            button === null || button === void 0 ? void 0 : button.setAttribute("disabled", "true");
        }
    }
    else {
        return;
    }
});
document.addEventListener("click", toggleSubmit);
function checkForm() {
    var inputs;
    if (signup) {
        inputs = document.getElementsByClassName("signup");
    }
    else {
        inputs = document.getElementsByClassName("login");
    }
    for (var i = 0; i < inputs.length; i++) {
        if (loading || inputs[i].value == "") {
            return false;
        }
    }
    var validEmail = !inputs[1].validity.patternMismatch;
    if (!validEmail || signup && inputs[2].value != inputs[3].value) {
        return false;
    }
    return true;
}
function toggleSubmit() {
    var button = document.querySelector("#submit");
    if (checkForm() == false) {
        button === null || button === void 0 ? void 0 : button.setAttribute("disabled", "true");
    }
    else {
        button === null || button === void 0 ? void 0 : button.removeAttribute("disabled");
    }
}
function postForm() {
    var msg = document.querySelector("p");
    var form = document.querySelector('form');
    var url = signup ? "/signup/form" : "/login/form";
    var formData = new FormData(form);
    var xhr = new XMLHttpRequest;
    xhr.open("POST", url);
    xhr.onloadstart = function () { loading = true; };
    xhr.onloadend = function () {
        msg.style.display = "block";
        switch (signup) {
            case false:
                if (xhr.status == 200) {
                    msg.innerText = xhr.response;
                    loading = false;
                    window.location.replace("/index");
                }
                else {
                    msg.innerText = "Invalid username or password";
                    setTimeout(function () {
                        loading = false;
                    }, 1000);
                }
                break;
            case true:
                if (xhr.status == 200) {
                    msg.innerText = "Sign Up Successful";
                    loading = false;
                    window.location.replace("/index");
                }
                else {
                    switch (xhr.response) {
                        case "mssql: Duplicate email":
                            msg.innerText = "Email is already taken";
                            break;
                        case "mssql: Duplicate uname":
                            msg.innerText = "Username is already taken";
                            break;
                        default:
                            msg.innerText = "Something went wrong. Try Again";
                    }
                    setTimeout(function () {
                        loading = false;
                    }, 1000);
                }
        }
    };
    xhr.send(formData);
}
function signupToggle() {
    signup = !signup;
    var btn = document.getElementById("submit");
    btn.value = signup ? "Sign Up" : "Login";
    var switchLink = document.getElementById("switch-link");
    switchLink.innerText = signup ? "Have an Account? Login" : "New? Create an Account";
    var inputs = document.getElementsByClassName("signup-only");
    for (var i = 0; i < inputs.length; i++) {
        if (!inputs[i].classList.contains("login")) {
            inputs[i].style.display = signup ? "flex" : "none";
        }
    }
    var allInputs = document.getElementsByClassName("signup");
    for (var i = 0; i < allInputs.length; i++) {
        if (allInputs[i]) {
            allInputs[i].value = "";
        }
    }
    var msg = document.querySelector("p");
    console.log(msg.style.display);
}
//# sourceMappingURL=form.js.map