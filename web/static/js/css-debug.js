console.log('CSS Variables Debug:');
const root = document.documentElement;
const computedStyle = getComputedStyle(root);

// Check CSS variables used by project dropdown
const cssVars = [
    '--bg-elevated',
    '--accent-primary', 
    '--border-radius',
    '--bg-secondary',
    '--border-light',
    '--text-primary',
    '--text-secondary',
    '--bg-hover',
    '--transition'
];

cssVars.forEach(varName => {
    const value = computedStyle.getPropertyValue(varName);
    console.log(`${varName}: ${value || 'UNDEFINED'}`);
});

// Check the actual dropdown element
setTimeout(() => {
    const dropdown = document.getElementById('project-dropdown');
    if (dropdown) {
        const dropdownStyle = getComputedStyle(dropdown);
        console.log('Project dropdown computed styles:');
        console.log('Display:', dropdownStyle.display);
        console.log('Visibility:', dropdownStyle.visibility);
        console.log('Opacity:', dropdownStyle.opacity);
        console.log('Position:', dropdownStyle.position);
        console.log('Z-index:', dropdownStyle.zIndex);
        console.log('Top:', dropdownStyle.top);
        console.log('Left:', dropdownStyle.left);
        console.log('Right:', dropdownStyle.right);
        console.log('Background:', dropdownStyle.backgroundColor);
        console.log('Border:', dropdownStyle.border);
        console.log('Width:', dropdownStyle.width);
        console.log('Height:', dropdownStyle.height);
        
        // Check if it has content
        console.log('Dropdown innerHTML length:', dropdown.innerHTML.length);
        console.log('Dropdown children count:', dropdown.children.length);
    } else {
        console.log('Dropdown element not found');
    }
}, 2000);

// Add a function to force dropdown visibility for testing
window.forceDropdownVisible = function() {
    const dropdown = document.getElementById('project-dropdown');
    if (dropdown) {
        dropdown.style.display = 'block';
        dropdown.style.opacity = '1';
        dropdown.style.visibility = 'visible';
        dropdown.style.zIndex = '10000';
        dropdown.style.backgroundColor = '#ffffff';
        dropdown.style.border = '2px solid red';
        dropdown.style.position = 'absolute';
        dropdown.style.top = '100%';
        dropdown.style.left = '0';
        dropdown.style.right = '0';
        dropdown.style.marginTop = '4px';
        console.log('Forced dropdown to be visible with red border');
    }
};

console.log('CSS debug script loaded. Use forceDropdownVisible() to test visibility.');
