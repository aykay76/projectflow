// Test script to verify header menu functionality
console.log('Testing header menu...');

// Wait for DOM to be ready
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM ready, checking elements...');
    
    const menuBtn = document.getElementById('header-menu-btn');
    const menu = document.getElementById('header-menu');
    
    console.log('Menu button:', menuBtn);
    console.log('Menu:', menu);
    
    if (menuBtn && menu) {
        console.log('Elements found, adding test click handler...');
        menuBtn.addEventListener('click', () => {
            console.log('TEST: Menu button clicked!');
            console.log('Menu display:', menu.style.display);
            console.log('Menu visibility:', menu.style.visibility);
            console.log('Menu classes:', menu.classList.toString());
        });
    } else {
        console.error('Missing elements!');
    }
});
