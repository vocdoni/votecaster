document.getElementById('aggregateButton').addEventListener('click', aggregateAddresses);

function aggregateAddresses() {
    const input = document.getElementById('fileInput');
    const files = input.files;
    if (files.length === 0) {
        alert('Please select one or more CSV files.');
        return;
    }

    const addresses = {};
    let filesProcessed = 0;

    Array.from(files).forEach(file => {
        const reader = new FileReader();
        reader.onload = function(e) {
            const text = e.target.result;
            processCsv(text);
            filesProcessed++;
            if (filesProcessed === files.length) {
                createDownloadLink(addresses);
            }
        };
        reader.readAsText(file);
    });

    function processCsv(csvText) {
        const lines = csvText.split('\n');
        lines.forEach(line => {
            const [address, weight] = line.split(',');
            if (address && weight) {
                addresses[address] = (addresses[address] || 0) + parseInt(weight, 10);
            }
        });
    }

    function createDownloadLink(data) {
        let csvContent = "data:text/csv;charset=utf-8,";
        for (const [address, weight] of Object.entries(data)) {
            csvContent += `${address},${weight}\n`;
        }
        
        const encodedUri = encodeURI(csvContent);
        const link = document.createElement('a');
        link.setAttribute('href', encodedUri);
        link.setAttribute('download', 'aggregated_addresses.csv');
        document.body.appendChild(link);
        
        document.getElementById('downloadButton').style.display = 'block';
        document.getElementById('downloadButton').onclick = function() {
            link.click();
        }
    }
}

