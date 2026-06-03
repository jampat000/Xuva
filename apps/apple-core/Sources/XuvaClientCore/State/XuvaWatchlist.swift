import Foundation
import SwiftUI

public struct WatchlistItem: Codable, Identifiable, Equatable {
    public var id: String
    public var kind: String
    public var title: String
    public var year: Int?
    public var posterUrl: String?
    public var backdropUrl: String?
    public var genres: [String]?
    public var addedAt: Date
}

@MainActor
public final class XuvaWatchlist: ObservableObject {
    @Published public private(set) var items: [WatchlistItem] = []

    private static let storageKey = "xuva.apple.watchlist"
    private var synced = false
    /// Identity of the API instance the cache was last synced against.
    /// Lets `syncFromServer()` re-run when a new pairing arrives.
    private var syncedFromBaseURL: URL?
    weak var api: XuvaAPI? {
        didSet {
            if api !== oldValue {
                synced = false
                if api == nil { reset() }
            }
        }
    }

    public init() {
        load()
    }

    /// Clear the local watchlist and its persisted cache. Called when the
    /// user unpairs so the next pairing doesn't inherit a stranger's items.
    public func reset() {
        items = []
        synced = false
        syncedFromBaseURL = nil
        UserDefaults.standard.removeObject(forKey: Self.storageKey)
    }

    public func isIn(id: String, kind: String) -> Bool {
        items.contains { $0.id == id && $0.kind == kind }
    }

    public func toggle(id: String, kind: String, title: String, year: Int?, posterUrl: String?, backdropUrl: String?, genres: [String]?) -> Bool {
        if isIn(id: id, kind: kind) {
            items.removeAll { $0.id == id && $0.kind == kind }
            persist()
            Task { try? await api?.watchlistRemove(mediaId: id, kind: kind) }
            return false
        } else {
            items.insert(WatchlistItem(id: id, kind: kind, title: title, year: year, posterUrl: posterUrl, backdropUrl: backdropUrl, genres: genres, addedAt: Date()), at: 0)
            persist()
            Task {
                let req = WatchlistAddRequest(mediaId: id, kind: kind, title: title, year: year, posterUrl: posterUrl, backdropUrl: backdropUrl, genres: genres)
                try? await api?.watchlistAdd(req)
            }
            return true
        }
    }

    /// Fetch the authoritative server list once per API instance and
    /// replace the local cache. Re-runs after a new pairing because
    /// `api`'s didSet clears `synced`.
    public func syncFromServer() async {
        guard let api else { return }
        if synced, syncedFromBaseURL == api.baseURL { return }
        do {
            let response = try await api.watchlistList()
            let formatter = ISO8601DateFormatter()
            items = response.items.compactMap { i in
                WatchlistItem(
                    id: i.mediaId,
                    kind: i.kind,
                    title: i.title,
                    year: i.year,
                    posterUrl: i.posterUrl,
                    backdropUrl: i.backdropUrl,
                    genres: i.genres,
                    addedAt: formatter.date(from: i.addedAt) ?? Date()
                )
            }
            persist()
            synced = true
            syncedFromBaseURL = api.baseURL
        } catch {
            xuvaLog("watchlist sync failed: \(error)")
        }
    }

    public func asHomeItems() -> [HomeItem] {
        items.map { entry in
            var item = HomeItem(
                id: entry.id,
                kind: entry.kind,
                title: entry.title,
                subtitle: entry.year.map(String.init),
                year: entry.year,
                genres: entry.genres,
                overview: nil
            )
            item.posterUrl = entry.posterUrl
            item.backdropUrl = entry.backdropUrl
            return item
        }
    }

    private func load() {
        guard let data = UserDefaults.standard.data(forKey: Self.storageKey),
              let decoded = try? JSONDecoder().decode([WatchlistItem].self, from: data) else { return }
        items = decoded
    }

    private func persist() {
        guard let data = try? JSONEncoder().encode(items) else { return }
        UserDefaults.standard.set(data, forKey: Self.storageKey)
    }
}
