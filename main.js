<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Streaming Client</title>
</head>
<body>
    <h1>Command Output</h1>
    <pre id="output"></pre>

    <script>
        const outputElement = document.getElementById('output');

        fetch('http://localhost:8080/command', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ cmd: "ls" })
        })
        .then(response => {
            const reader = response.body.getReader();
            reader.read().then(function processText({ done, value }) {
                if (done) {
                    console.log("Stream complete");
                    return;
                }
                const text = new TextDecoder().decode(value);
                console.log(text); // Log to console
                outputElement.textContent += text; // Display on the page
                reader.read().then(processText);
            });
        })
        .catch(err => console.error('Fetch error:', err));
    </script>
</body>
</html>

  