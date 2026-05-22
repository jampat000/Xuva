import Foundation
import SwiftUI

/// Local watchlist mirroring the web's `apps/web/svelte/src/lib/stores/watchlistStore.svelte.ts` —
/// client-side only, persisted to UserDefaults. When the server gains a real
/// watchlist API later this can switch to an HTTP-backed implementation
/// without changing the call sites.
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

    public init() {
        load()
    }

    public func isIn(id: String, kind: String) -> Bool {
        items.contains { $0.id == id && $0.kind == kind }
    }

    public func toggle(id: String, kind: String, title: String, year: Int?, posterUrl: String?, backdropUrl: String?, genres: [String]?) -> Bool {
        if isIn(id: id, kind: kind) {
            items.removeAll { $0.id == id && $0.kind == kind }
            persist()
            return false
        } else {
            items.insert(WatchlistItem(id: id, kind: kind, title: title, year: year, posterUrl: posterUrl, backdropUrl: backdropUrl, genres: genres, addedAt: Date()), at: 0)
            persist()
            return true
        }
    }

    public func asHomeItems() -> [HomeItem] {
        items.map { entry in
            HomeItem(
                id: entry.id,
                kind: entry.kind,
                title: entry.title,
                subtitle: entry.year.map(String.init),
                year: entry.year,
                genres: entry.genres,
                overview: nil
            )
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
