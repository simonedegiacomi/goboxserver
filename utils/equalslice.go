package utils

// Function used to compare the password hash
func ComparePassword(a, b []byte) bool {
    // This will never appen, because the password
    // are hashed, so their length are always the same,
    // but just to be sure...
    if len(a) != len(b) {
        return false;
    }
    
    // Check if every byte is he same in the slices
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    
    // Yep, they'r equal
    return true
}