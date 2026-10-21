
const form = document.querySelector("#shorten-form");
const input = document.querySelector("#url-input");
const shortenButton = document.querySelector("#shorten-button");
const buttonText = document.querySelector("#button-text");

const errorMessage = document.querySelector("#error-message");
const result = document.querySelector("#result");
const shortURL = document.querySelector("#short-url");
const targetURL = document.querySelector("#target-url");
const copyButton = document.querySelector("#copy-button");
const copyMessage = document.querySelector("#copy-message");

form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const url = input.value.trim();

    errorMessage.hidden = true;
    result.hidden = true;
    copyMessage.textContent = "";

    shortenButton.disabled = true;
    buttonText.textContent = "Shortening...";

    try {
        const response = await fetch("/shorten", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ url }),
        });

        if (!response.ok) {
            const message = await response.text();
            throw new Error(message || "Could not shorten this URL.");
        }

        const data = await response.json();

        shortURL.href = data.short_url;
        shortURL.textContent = data.short_url;
        targetURL.textContent = data.target_url;

        result.hidden = false;
    } catch (error) {
        errorMessage.textContent =
            error.message || "Something went wrong. Please try again.";
        errorMessage.hidden = false;
    } finally {
        shortenButton.disabled = false;
        buttonText.textContent = "Shorten URL";
    }
});

copyButton.addEventListener("click", async () => {
    try {
        await navigator.clipboard.writeText(shortURL.href);
        copyMessage.textContent = "Copied to clipboard!";
        copyButton.textContent = "Copied ✓";
    } catch {
        copyMessage.textContent =
            "Copy wasn't available. Select the link and copy it manually.";
    }
});
