<script>
    export let size = "4rem";
</script>   

<div class="loading-container">
    <div class="spinner" style="width: {size}; height: {size};"></div>
</div>

<style>
  .loading-container{
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #fff;
    text-align: center;
    height: 100%;
  }

  .spinner {
    margin-bottom: 1rem;
    border: 4px solid #ffffff33;
    border-top: 4px solid #ffffff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>