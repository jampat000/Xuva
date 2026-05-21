//
//  Item.swift
//  Xuva
//
//  Created by James on 21/5/2026.
//

import Foundation
import SwiftData

@Model
final class Item {
    var timestamp: Date

    init(timestamp: Date) {
        self.timestamp = timestamp
    }
}
