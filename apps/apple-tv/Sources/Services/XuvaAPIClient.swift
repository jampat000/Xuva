import Foundation

enum XuvaAPIError: Error, LocalizedError {
    case invalidBaseURL
    case invalidResponse
    case server(Int)

    var errorDescription: String? {
        switch self {
        case .invalidBaseURL:
            return "Server URL is not valid."
        case .invalidResponse:
            return "Xuva returned data this app could not read."
        case .server(let status):
            return "Xuva returned HTTP \(status)."
        }
    }
}

final class XuvaAPIClient {
    private let baseURL: URL
    private let session: URLSession

    init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    convenience init(serverURL: String) throws {
        guard let baseURL = URL(string: serverURL.trimmingCharacters(in: .whitespacesAndNewlines)) else {
            throw XuvaAPIError.invalidBaseURL
        }
        self.init(baseURL: baseURL)
    }

    func bootstrap(clientProfile: String = "apple-tv") async throws -> ClientBootstrap {
        var components = URLComponents(url: baseURL.appending(path: "/api/client/bootstrap"), resolvingAgainstBaseURL: false)
        components?.queryItems = [URLQueryItem(name: "clientProfile", value: clientProfile)]
        guard let url = components?.url else {
            throw XuvaAPIError.invalidBaseURL
        }
        return try await get(url)
    }

    func createPairingRequest(deviceName: String, clientProfile: String = "apple-tv") async throws -> PairingRequest {
        let url = baseURL.appending(path: "/api/pairing/requests")
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(PairingRequestPayload(deviceName: deviceName, clientProfile: clientProfile))
        return try await send(request)
    }

    func pairingStatus(id: String) async throws -> PairingRequest {
        try await get(baseURL.appending(path: "/api/pairing/requests/\(id)"))
    }

    func home(clientProfile: String = "apple-tv") async throws -> TVHomeResponse {
        var components = URLComponents(url: baseURL.appending(path: "/api/client/home"), resolvingAgainstBaseURL: false)
        components?.queryItems = [URLQueryItem(name: "clientProfile", value: clientProfile)]
        guard let url = components?.url else {
            throw XuvaAPIError.invalidBaseURL
        }
        return try await get(url)
    }

    private func get<T: Decodable>(_ url: URL) async throws -> T {
        let (data, response) = try await session.data(from: url)
        return try decode(data: data, response: response)
    }

    private func send<T: Decodable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)
        return try decode(data: data, response: response)
    }

    private func decode<T: Decodable>(data: Data, response: URLResponse) throws -> T {
        guard let http = response as? HTTPURLResponse else {
            throw XuvaAPIError.invalidResponse
        }
        guard (200..<300).contains(http.statusCode) else {
            throw XuvaAPIError.server(http.statusCode)
        }
        return try JSONDecoder().decode(T.self, from: data)
    }
}
