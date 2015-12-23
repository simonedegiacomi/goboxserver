package utils

func ComparePassword(a, b []byte) bool {
    if len(a) != len(b) {
        return false;
    }
    
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    // Yep, they'r equal
    return true
}