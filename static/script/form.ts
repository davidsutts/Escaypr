// Global variables.
var loading = false;
var signup = false;

// Check for changes on form to update submission button.
document.addEventListener("DOMContentLoaded", () => {
	document.getElementById("pword")?.addEventListener("change", toggleSubmit);
	document.getElementById("uname")?.addEventListener("change", toggleSubmit);
})

// Submit form when enter is pressed.
document.addEventListener("keyup", function (event: KeyboardEvent) {
	var button = document.querySelector<HTMLElement>("#submit");
	toggleSubmit();
	if (event.code == "Enter") {
		if (checkForm()) {
			postForm();
			button?.setAttribute("disabled", "true");
		}
	} else {
		return;
	}
});

// Update submit button whenever user interacts with page.
document.addEventListener("click", toggleSubmit);

// Check the form to ensure the required fields are not blank. If in signup mode, check that the 
// passwords match.
function checkForm(): boolean {
	let inputs: HTMLCollectionOf<HTMLInputElement>;
	if (signup) {
		inputs = document.getElementsByClassName("signup") as HTMLCollectionOf<HTMLInputElement>;
	} else {
		inputs = document.getElementsByClassName("login") as HTMLCollectionOf<HTMLInputElement>;
	}
	for (var i = 0; i < inputs.length; i++) {
		if (loading || inputs[i].value == "") {
			return false;
		}
	}
	let validEmail = !inputs[1].validity.valid;
	if (!validEmail || signup && inputs[2].value != inputs[3].value) {
		return false;
	}

	return true;
}

// Toggle the submit button depending on the checkForm return.
function toggleSubmit() {
	var button = document.querySelector<HTMLElement>("#submit");
	if (checkForm() == false) {
		button?.setAttribute("disabled", "true");
	} else {
		button?.removeAttribute("disabled");
	}
}

// Post the inputs to the form.
function postForm() {
	const msg = document.querySelector("p")!;
	const form = document.querySelector('form')!;
	const url = signup ? "/signup/form" : "/login/form";

	const formData = new FormData(form);

	let xhr = new XMLHttpRequest;
	xhr.open("POST", url);

	// Load response and redirect or fail attempt.
	xhr.onloadstart = () => { loading = true; }
	xhr.onloadend = () => {
		msg.style.display = "block";
		switch (signup) {
			case false:
				if (xhr.status == 200) {
					msg.innerText = xhr.response;
					loading = false;
					window.location.replace("/");	// Send logged in user to home.
				} else {
					msg.innerText = "Invalid username or password";
					setTimeout(() => {
						loading = false;
					}, 1000);	// Make user wait before trying again.
				}
				break
			case true:
				if (xhr.status == 200) {
					msg.innerText = "Sign Up Successful";
					loading = false;
					window.location.replace("/");	// Send logged in user to home.
				} else {
					switch (xhr.response) {
						case "duplicate key err":
							msg.innerText = "Username or Email is already taken";
							break
						default:
							msg.innerText = "Something went wrong. Try Again"
					}
					setTimeout(() => {
						loading = false;
					}, 1000);	// Make user wait before trying again.
				}

		}

	}

	// Send login/signup request.
	xhr.send(formData);

}

// Switch the user to the signup page.
function signupToggle() {
	// Toggle signup var.
	signup = !signup

	// Replace text in button.
	let btn = document.getElementById("submit") as HTMLInputElement;
	btn.value = signup ? "Sign Up" : "Login"

	// Replace tool tip in link.
	let switchLink = document.getElementById("switch-link") as HTMLElement;
	switchLink.innerText = signup ? "Have an Account? Login" : "New? Create an Account"

	// Get signup specific fields and toggle visibility.
	let inputs = document.getElementsByClassName("signup-only") as HTMLCollectionOf<HTMLDivElement>
	for (let i = 0; i < inputs.length; i++) {
		if (!inputs[i].classList.contains("login")) {
			inputs[i].style.display = signup ? "flex" : "none"
		}
	}

	// Clear values of all sign-up fields.
	let allInputs = document.getElementsByClassName("signup") as HTMLCollectionOf<HTMLInputElement>
	for (let i = 0; i < allInputs.length; i++) {
		if (allInputs[i]) {
			allInputs[i].value = ""
		}
	}

	// Clear error message.
	const msg = document.querySelector("p")!;
	console.log(msg.style.display)

}