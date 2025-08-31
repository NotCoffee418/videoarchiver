<script>
  let {
    onAccept = () => {}
  } = $props();

  function handleAccept() {
    onAccept();
  }

  function handleDecline() {
    // Close the application on decline
    window.close();
  }
</script>

<div class="disclaimer-overlay">
  <div class="disclaimer-content">
    <h1>Legal Disclaimer</h1>
    <div class="disclaimer-text">
      <p>By using this application, you agree to the following terms:</p>
      <ul>
        <li>You will only use this application for <strong>legal purposes</strong> in accordance with the laws of your country and jurisdiction.</li>
        <li>You are responsible for ensuring that any content you download is legally obtained and that you have the right to download and store such content.</li>
        <li>This application is provided as-is, and the developers are not responsible for any misuse of the software.</li>
        <li>You understand that downloading copyrighted content without permission may violate copyright laws in your jurisdiction.</li>
        <li>It is your responsibility to comply with all applicable laws, including but not limited to copyright laws, privacy laws, and terms of service of third-party platforms.</li>
      </ul>
      <p>By clicking "I Agree", you acknowledge that you have read, understood, and agree to these terms and will use this application only for legal purposes.</p>
    </div>
    <div class="button-container">
      <button class="decline-btn" onclick={handleDecline}>I Decline</button>
      <button class="accept-btn" onclick={handleAccept}>I Agree</button>
    </div>
  </div>
</div>

<style>
  .disclaimer-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.9);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
  }

  .disclaimer-content {
    background-color: #1e1e1e;
    color: #ffffff;
    padding: 2rem;
    border-radius: 12px;
    max-width: 600px;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    border: 1px solid #333;
  }

  h1 {
    color: #ff6b6b;
    text-align: center;
    margin-bottom: 1.5rem;
    font-size: 1.8rem;
  }

  .disclaimer-text {
    line-height: 1.6;
    margin-bottom: 2rem;
  }

  .disclaimer-text p {
    margin-bottom: 1rem;
  }

  .disclaimer-text ul {
    margin: 1rem 0;
    padding-left: 1.5rem;
  }

  .disclaimer-text li {
    margin-bottom: 0.8rem;
  }

  .disclaimer-text strong {
    color: #ff6b6b;
  }

  .button-container {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
  }

  .accept-btn, .decline-btn {
    flex: 1;
    padding: 12px 24px;
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .accept-btn {
    background-color: #28a745;
    color: white;
  }

  .accept-btn:hover {
    background-color: #218838;
  }

  .decline-btn {
    background-color: #dc3545;
    color: white;
  }

  .decline-btn:hover {
    background-color: #c82333;
  }

  /* Scrollbar styling for webkit browsers */
  .disclaimer-content::-webkit-scrollbar {
    width: 8px;
  }

  .disclaimer-content::-webkit-scrollbar-track {
    background: #2c2c2c;
    border-radius: 4px;
  }

  .disclaimer-content::-webkit-scrollbar-thumb {
    background: #555;
    border-radius: 4px;
  }

  .disclaimer-content::-webkit-scrollbar-thumb:hover {
    background: #777;
  }
</style>