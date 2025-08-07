async function fetchOrder() {
    const orderUid = document.getElementById('orderUid').value.trim();
    const resultDiv = document.getElementById('result');
    const errorDiv = document.getElementById('error');

    // Очистка предыдущих результатов
    resultDiv.innerHTML = '';
    errorDiv.innerHTML = '';

    if (!orderUid) {
        errorDiv.innerHTML = 'Please enter an Order UID';
        return;
    }

    try {
        const response = await fetch(`http://localhost:8080/order/${orderUid}`);
        if (!response.ok) {
            if (response.status === 404) {
                errorDiv.innerHTML = `Order ${orderUid} not found`;
            } else {
                errorDiv.innerHTML = `Error: ${response.status} ${response.statusText}`;
            }
            return;
        }

        const order = await response.json();
        // Форматирование JSON с отступами
        resultDiv.innerHTML = `<pre>${JSON.stringify(order, null, 2)}</pre>`;
    } catch (error) {
        errorDiv.innerHTML = `Failed to fetch order: ${error.message}`;
    }
}