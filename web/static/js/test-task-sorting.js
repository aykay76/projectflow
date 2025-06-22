/**
 * Test script for numeric task sorting functions
 * Run this in browser console to test the sorting logic
 */

// Import functions (for testing in browser, these would be available globally)
// import { extractTaskNumber, compareTasksByNumber, sortTasksByNumber } from './utils.js';

// Test data
const testTasks = [
    { id: '1', display_id: 'PF-5', title: 'Task 5' },
    { id: '2', display_id: 'PF-123', title: 'Task 123' },
    { id: '3', display_id: 'PF-1', title: 'Task 1' },
    { id: '4', display_id: 'PF-45', title: 'Task 45' },
    { id: '5', display_id: 'ABC-2', title: 'ABC Task 2' },
    { id: '6', display_id: 'ABC-100', title: 'ABC Task 100' },
    { id: '7', display_id: 'PF-10', title: 'Task 10' },
    { id: '8', display_id: null, title: 'Task without ID' },
    { id: '9', display_id: 'INVALID', title: 'Invalid ID format' }
];

function runTests() {
    console.log('Testing extractTaskNumber function:');
    console.log('PF-123 ->', extractTaskNumber('PF-123')); // Should be 123
    console.log('ABC-456 ->', extractTaskNumber('ABC-456')); // Should be 456
    console.log('PF-1 ->', extractTaskNumber('PF-1')); // Should be 1
    console.log('null ->', extractTaskNumber(null)); // Should be 0
    console.log('INVALID ->', extractTaskNumber('INVALID')); // Should be 0
    console.log('');
    
    console.log('Original task order:');
    testTasks.forEach(task => console.log(`${task.display_id || 'NO_ID'}: ${task.title}`));
    console.log('');
    
    console.log('Sorted task order:');
    const sorted = sortTasksByNumber(testTasks);
    sorted.forEach(task => console.log(`${task.display_id || 'NO_ID'}: ${task.title}`));
    console.log('');
    
    console.log('Expected order: NO_ID, INVALID, PF-1, ABC-2, PF-5, PF-10, PF-45, ABC-100, PF-123');
}

// Export for use in browser testing
if (typeof window !== 'undefined') {
    window.testTaskSorting = runTests;
}
