// Global variables.
var loading = false;
// Check for changes on form to update submission button.
document.addEventListener("DOMContentLoaded", function () {
    var _a, _b;
    (_a = document.getElementById("pword")) === null || _a === void 0 ? void 0 : _a.addEventListener("change", toggleSubmit);
    (_b = document.getElementById("uname")) === null || _b === void 0 ? void 0 : _b.addEventListener("change", toggleSubmit);
});
// Submit form when enter is pressed.
document.addEventListener("keyup", function (event) {
    var button = document.querySelector("#submit");
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
// Check the form to ensure the username and password are not blank
function checkForm() {
    var inputs = document.querySelectorAll("input");
    for (var i = 0; i < 2; i++) {
        if (loading || inputs[i].value == "") {
            return false;
        }
    }
    return true;
}
// Toggle the submit button depending on the checkForm return.
function toggleSubmit() {
    var button = document.querySelector("#submit");
    if (checkForm() == false) {
        button === null || button === void 0 ? void 0 : button.setAttribute("disabled", "true");
    }
    else {
        button === null || button === void 0 ? void 0 : button.removeAttribute("disabled");
    }
}
// Post the inputs to the form.
function postForm() {
    var msg = document.querySelector("p");
    var form = document.querySelector('form');
    var url = "/login/form";
    var formData = new FormData(form);
    var xhr = new XMLHttpRequest;
    xhr.open("POST", url);
    // Load response and redirect or fail attempt.
    xhr.onloadstart = function () { loading = true; };
    xhr.onloadend = function () {
        msg.style.display = "block";
        if (xhr.status == 200) {
            msg.innerText = "Login Successful";
            loading = false;
            window.location.replace("/index"); // Send logged in user to index.
        }
        else {
            msg.innerText = "Invalid usernamne or password";
            setTimeout(function () {
                loading = false;
            }, 1000); // Make user wait before trying again.
        }
    };
    // Send login request.
    xhr.send(formData);
}
