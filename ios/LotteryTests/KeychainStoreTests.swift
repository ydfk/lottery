import Testing
@testable import Lottery

@Suite(.serialized)
struct KeychainStoreTests {
    @Test("JWT 可以安全保存和删除")
    func tokenLifecycle() throws {
        let store = KeychainStore()
        store.deleteToken()
        try store.saveToken("test-token")
        #expect(store.readToken() == "test-token")
        store.deleteToken()
        #expect(store.readToken() == nil)
    }
}

